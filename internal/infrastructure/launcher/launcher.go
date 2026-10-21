// Package launcher связывает владельца, controller и независимый recovery без RPC-схемы.
package launcher

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"wireglow/internal/application/controller"
	"wireglow/internal/application/ports"
	"wireglow/internal/backend"
	"wireglow/internal/infrastructure/network"
	"wireglow/internal/infrastructure/persistence"
	"wireglow/internal/infrastructure/platform"
	"wireglow/internal/rpc"
)

type Request struct {
	Version   int               `json:"version"`
	RunID     string            `json:"runId"`
	Owner     platform.Identity `json:"owner"`
	Options   backend.Options   `json:"options"`
	ProfileID string            `json:"profileId"`
	Desktop   bool              `json:"desktop,omitempty"`
}
type Lease struct {
	RunID    string `json:"runId"`
	Sequence uint64 `json:"sequence"`
}
type Report struct {
	Snapshot      controller.Snapshot `json:"snapshot"`
	Exited        bool                `json:"exited"`
	CleanupFailed bool                `json:"cleanupFailed"`
	SessionFailed bool                `json:"sessionFailed"`
}

func readRequest(dir string) (Request, error) {
	var r Request
	if !filepath.IsAbs(dir) {
		return r, errors.New("launcher: absolute run directory required")
	}
	if err := persistence.ReadJSON(filepath.Join(dir, "request.json"), &r); err != nil {
		return r, err
	}
	if r.Version != 1 || r.RunID == "" {
		return r, errors.New("launcher: invalid request")
	}
	return r, nil
}

// Endpoint передаётся владельцу по закрытому каналу запуска GUI.
type Endpoint struct {
	Socket string `json:"socket"`
	Token  string `json:"token"`
}

// RunDesktop сохраняет владельца и recovery, а команды принимает через gRPC.
func RunDesktop(ctx context.Context, o backend.Options, ready func(Endpoint)) error {
	return run(ctx, o, "", nil, ready)
}

// Run держит Job на Windows; Linux держит службу, а controller контролирует lease владельца.
func Run(ctx context.Context, o backend.Options, id string, onSnapshot func(controller.Snapshot)) error {
	return run(ctx, o, id, onSnapshot, nil)
}
func run(ctx context.Context, o backend.Options, id string, onSnapshot func(controller.Snapshot), ready func(Endpoint)) error {
	// Вывод frontend не задерживает lease владельца и остановку процессов.
	updates := make(chan controller.Snapshot, 1)
	defer close(updates)
	go func() {
		for snapshot := range updates {
			if onSnapshot != nil {
				onSnapshot(snapshot)
			}
		}
	}()
	if err := platform.EnsurePaths(o.Paths); err != nil {
		return err
	}
	ownerLock, err := platform.AcquireLock(filepath.Join(o.Paths.Runtime, "launcher.lock"))
	if err != nil {
		return err
	}
	defer ownerLock.Close()
	identity, err := platform.ProcessIdentity(os.Getpid())
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	runID := hex.EncodeToString(bytes)
	dir, err := os.MkdirTemp(o.Paths.Runtime, "owner-*")
	if err != nil {
		return err
	}
	// Run-directory остаётся до конца watchdog: при crash это его источник identity.
	r := Request{Version: 1, RunID: runID, Owner: identity, Options: o, ProfileID: id, Desktop: ready != nil}
	if err := persistence.WriteJSON(filepath.Join(dir, "request.json"), r); err != nil {
		return err
	}
	if err := persistence.WriteJSON(filepath.Join(dir, "lease.json"), Lease{runID, 1}); err != nil {
		return err
	}
	watcher, err := platform.StartRecovery(ctx, ports.Command{Path: executable, Args: []string{"watchdog", "--run-dir", dir}}, "wireglow-recovery-"+runID)
	if err != nil {
		return err
	}
	startupCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	err = awaitFile(startupCtx, filepath.Join(dir, "watchdog-ready"), watcher)
	cancel()
	if err != nil {
		stopProcess(watcher)
		return err
	}
	worker, err := platform.StartController(ctx, ports.Command{Path: executable, Args: []string{"worker", "--run-dir", dir}}, "wireglow-controller-"+runID)
	if err != nil {
		stopProcess(watcher)
		return err
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	sequence := uint64(1)
	var result error
	var last controller.Snapshot
	announced := false
	for {
		select {
		case <-ctx.Done():
			// Остановка обновления lease вызывает общий cleanup; force stop ограничен сроком.
			goto finish
		case <-worker.Done():
			result = worker.Err()
			goto finish
		case <-watcher.Done():
			result = errors.New("launcher: recovery watchdog exited")
			goto finish
		case <-ticker.C:
			if ready != nil && !announced {
				var endpoint Endpoint
				if persistence.ReadJSON(filepath.Join(dir, "rpc.json"), &endpoint) == nil {
					ready(endpoint)
					announced = true
				}
			}
			sequence++
			if err := persistence.WriteJSON(filepath.Join(dir, "lease.json"), Lease{runID, sequence}); err != nil {
				result = err
				goto finish
			}
			var report Report
			if err := persistence.ReadJSON(filepath.Join(dir, "report.json"), &report); err == nil {
				if report.Snapshot != last {
					last = report.Snapshot
					select {
					case <-updates:
					default:
					}
					updates <- last
				}
				if report.Exited {
					if report.CleanupFailed || report.SessionFailed || report.Snapshot.State == controller.Failed {
						result = errors.New("launcher: controller session failed")
					}
					goto finish
				}
			}
		}
	}
finish:
	// Закрытый lease немедленно отменяет startup; файл не содержит profile/ключи.
	_ = os.Remove(filepath.Join(dir, "lease.json"))
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 30*time.Second)
	select {
	case <-worker.Done():
	case <-waitCtx.Done():
		result = errors.Join(result, stopProcess(worker))
	}
	waitCancel()
	// Watchdog должен закончить recovery прежде, чем launcher освободит свою блокировку.
	waitCtx, waitCancel = context.WithTimeout(context.Background(), 30*time.Second)
	select {
	case <-watcher.Done():
		result = errors.Join(result, watcher.Err())
	case <-waitCtx.Done():
		result = errors.Join(result, errors.New("launcher: recovery still pending"))
	}
	waitCancel()
	var report Report
	if err := persistence.ReadJSON(filepath.Join(dir, "report.json"), &report); err == nil && report.CleanupFailed {
		result = errors.Join(result, errors.New("launcher: cleanup incomplete"))
	}
	select {
	case <-watcher.Done():
		select {
		case <-worker.Done():
			_ = os.RemoveAll(dir)
		default:
		}
	default:
	}
	return result
}

func stopProcess(p ports.Process) error {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	return p.Stop(ctx)
}
func awaitFile(ctx context.Context, path string, p ports.Process) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return errors.New("launcher: readiness timed out")
		case <-p.Done():
			return errors.New("launcher: process exited before readiness")
		case <-ticker.C:
		}
	}
}

// Worker — entrypoint службы или Job; desktop-команды принимает локальный gRPC.
func Worker(ctx context.Context, dir string) error {
	if err := platform.RequireManagedController(); err != nil {
		return err
	}
	r, err := readRequest(dir)
	if err != nil {
		return err
	}
	identity, err := platform.ProcessIdentity(os.Getpid())
	if err != nil {
		return err
	}
	if err := persistence.WriteJSON(filepath.Join(dir, "controller.json"), identity); err != nil {
		return err
	}
	if err := persistence.ReadJSON(filepath.Join(dir, "watchdog.json"), &r.Options.RecoveryIdentity); err != nil {
		return err
	}
	b, err := backend.Open(ctx, r.Options)
	if err != nil {
		return err
	}
	defer b.Close()
	beats := make(chan struct{}, 1)
	leaseCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go feedOwner(leaseCtx, r, dir, beats)
	ownerDone := make(chan error, 1)
	go func() { ownerDone <- b.Controller.WatchOwner(ctx, beats, 8*time.Second, platform.AwakeTime) }()
	events, unsubscribe := b.Controller.Subscribe()
	defer unsubscribe()
	reportsDone := make(chan struct{})
	go func() {
		defer close(reportsDone)
		for snapshot := range events {
			_ = persistence.WriteJSON(filepath.Join(dir, "report.json"), Report{Snapshot: snapshot})
		}
	}()
	if r.Desktop {
		token := make([]byte, 32)
		if _, err = rand.Read(token); err != nil {
			return err
		}
		socket, socketErr := platform.DesktopSocket(dir, r.RunID)
		if socketErr != nil {
			return socketErr
		}
		defer platform.CleanupDesktopSocket(r.RunID)
		endpoint := Endpoint{Socket: socket, Token: hex.EncodeToString(token)}
		listener, listenErr := net.Listen("unix", endpoint.Socket)
		if listenErr != nil {
			return listenErr
		}
		defer listener.Close()
		// Windows sockets inherit the user/SYSTEM-only DACL of the private
		// runtime directory. chmod is POSIX-only here: Windows can fail while
		// opening the AF_UNIX reparse point to change its attributes.
		if runtime.GOOS != "windows" {
			if err = os.Chmod(endpoint.Socket, 0600); err != nil {
				return err
			}
		}
		server, serverErr := rpc.NewServer(b, rpc.Options{Context: leaseCtx, Token: token, BuildID: "0.1.0-dev", Epoch: r.RunID})
		if serverErr != nil {
			return serverErr
		}
		defer server.Stop()
		go func() {
			if server.Serve(listener) != nil {
				cancel()
			}
		}()
		if err = persistence.WriteJSON(filepath.Join(dir, "rpc.json"), endpoint); err != nil {
			return err
		}
	} else {
		err = b.Connect(ctx, r.ProfileID)
	}
	if err == nil {
		for {
			select {
			case ownerErr := <-ownerDone:
				err = ownerErr
				goto done
			case <-leaseCtx.Done():
				goto done
			case <-time.After(250 * time.Millisecond):
				if !r.Desktop && b.Controller.Snapshot().State == controller.Failed {
					err = errors.New("controller: session failed")
					goto done
				}
			}
		}
	}
done:
	snapshot := b.Controller.Snapshot()
	closeErr := b.Close()
	cancel()
	unsubscribe()
	<-reportsDone
	if err == nil {
		snapshot = b.Controller.Snapshot()
	}
	_ = persistence.WriteJSON(filepath.Join(dir, "report.json"), Report{Snapshot: snapshot, Exited: true, CleanupFailed: closeErr != nil, SessionFailed: err != nil})
	return errors.Join(err, closeErr)
}

func feedOwner(ctx context.Context, r Request, dir string, beats chan<- struct{}) {
	defer close(beats)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var sequence uint64
	for {
		alive, err := r.Owner.Alive()
		if err != nil || !alive {
			return
		}
		var lease Lease
		if err := persistence.ReadJSON(filepath.Join(dir, "lease.json"), &lease); err != nil || lease.RunID != r.RunID {
			return
		}
		if lease.Sequence > sequence {
			sequence = lease.Sequence
			select {
			case beats <- struct{}{}:
			default:
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// Watchdog запускается отдельно от группы controller и до применения системного proxy.
func Watchdog(ctx context.Context, dir string) error {
	r, err := readRequest(dir)
	if err != nil {
		return err
	}
	if err := platform.EnsurePaths(r.Options.Paths); err != nil {
		return err
	}
	identitySelf, err := platform.ProcessIdentity(os.Getpid())
	if err != nil {
		return err
	}
	if err := persistence.WriteJSON(filepath.Join(dir, "watchdog.json"), identitySelf); err != nil {
		return err
	}
	if err := persistence.WriteAtomic(filepath.Join(dir, "watchdog-ready"), []byte("ready\n")); err != nil {
		return err
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	started := time.Now()
	var identity platform.Identity
	for {
		if err := persistence.ReadJSON(filepath.Join(dir, "controller.json"), &identity); err == nil {
			break
		}
		alive, err := r.Owner.Alive()
		if err != nil {
			return err
		}
		if !alive || time.Since(started) > 30*time.Second {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
	for {
		alive, err := identity.Alive()
		if err != nil {
			return err
		}
		if !alive {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
	// Новый controller не может одновременно менять настройки во время recovery.
	lockCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	var lock *platform.Lock
	for {
		lock, err = platform.AcquireLock(filepath.Join(r.Options.Paths.Runtime, "controller.lock"))
		if err == nil {
			break
		}
		if !errors.Is(err, platform.ErrLocked) {
			return err
		}
		select {
		case <-lockCtx.Done():
			return errors.New("recovery: controller lock timeout")
		case <-ticker.C:
		}
	}
	defer lock.Close()
	manager := &network.Manager{Path: filepath.Join(r.Options.Paths.Data, "proxy-journal.json"), Settings: platform.ProxySettings{}}
	recoveryCtx, recoveryCancel := context.WithTimeout(ctx, 20*time.Second)
	defer recoveryCancel()
	var ipcErr error
	if r.Desktop {
		ipcErr = platform.CleanupDesktopSocket(r.RunID)
	}
	return errors.Join(manager.Recover(recoveryCtx), persistence.CleanupSession(r.Options.Paths.Runtime), ipcErr)
}
