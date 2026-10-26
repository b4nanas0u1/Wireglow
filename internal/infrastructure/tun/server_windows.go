package tun

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	winio "github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
	"wireglow/internal/application/ports"
	"wireglow/internal/infrastructure/platform"
)

type windowsJournal struct {
	Version int    `json:"version"`
	Device  string `json:"device"`
	Stage   string `json:"stage"`
}

type windowsSession struct {
	journal     windowsJournal
	journalPath string
	stage       string
	owner       windowsOwner
	uplink      windowsLink
	process     ports.Process
	started     bool
	cleaned     bool
	idle        *time.Timer
}

func randomWindowsID() (string, error) {
	var b [16]byte
	_, err := rand.Read(b[:])
	return hex.EncodeToString(b[:]), err
}

// The UAC-approved, already mapped Go executable remains alive. Only pinned
// mihomo is copied to a protected directory. Its Wintun DLL is embedded in the
// pinned executable; no DLL is loaded from the portable directory.
func RunWindowsPortableHelper(ctx context.Context, ownerPID int, ownerStart string) (result error) {
	// Windows mutex ownership belongs to a thread, not a goroutine.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if !windows.GetCurrentProcessToken().IsElevated() {
		return errors.New("tun: administrator elevation is required")
	}
	owner, err := processOwner(ownerPID)
	if err != nil || owner.Identity.Start != ownerStart {
		return errors.New("tun: controller identity changed")
	}
	if !strings.EqualFold(filepath.Base(owner.Identity.Path), "controller.exe") {
		return errors.New("tun: unexpected controller executable")
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	if !strings.EqualFold(filepath.Dir(self), filepath.Dir(owner.Identity.Path)) {
		return errors.New("tun: helper and controller must belong to the same bundle")
	}
	root := filepath.Dir(filepath.Dir(owner.Identity.Path))
	if info, err := os.Lstat(filepath.Join(root, "portable.json")); err != nil || !info.Mode().IsRegular() {
		return errors.New("tun: portable bundle required")
	}
	// A kernel mutex provides machine-wide exclusion without moving the portable
	// recovery journal or lock files into the system profile.
	sd, err := windows.SecurityDescriptorFromString("D:P(A;;GA;;;SY)(A;;GA;;;BA)")
	if err != nil {
		return err
	}
	name, _ := windows.UTF16PtrFromString(`Global\Wireglow-Tun-Session-v1`)
	mutex, err := windows.CreateMutex(&windows.SecurityAttributes{Length: 24, SecurityDescriptor: sd}, false, name)
	if err != nil && !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return err
	}
	defer windows.CloseHandle(mutex)
	wait, err := windows.WaitForSingleObject(mutex, 0)
	if err != nil || (wait != windows.WAIT_OBJECT_0 && wait != windows.WAIT_ABANDONED) {
		return errors.New("tun: another helper owns the network session")
	}
	defer windows.ReleaseMutex(mutex)
	journalDir := filepath.Join(root, "data", "runtime", "tun")
	unpin, err := pinWindowsDirectories(filepath.Dir(journalDir))
	if err != nil {
		return err
	}
	defer unpin()
	if err = adminDirectory(journalDir); err != nil {
		return err
	}
	unpinJournal, err := pinWindowsDirectories(journalDir)
	if err != nil {
		return err
	}
	defer unpinJournal()
	if err = adminDirectory(journalDir); err != nil {
		return err
	}
	journalPath := filepath.Join(journalDir, "journal.json")
	recoveryCtx, recoveryCancel := context.WithTimeout(ctx, 12*time.Second)
	err = recoverWindowsJournal(recoveryCtx, journalPath)
	recoveryCancel()
	if err != nil {
		return err
	}
	uplink, err := windowsNetwork("")
	if err != nil {
		return err
	}
	id, err := randomWindowsID()
	if err != nil {
		return err
	}
	stages, err := stageRoot()
	if err != nil {
		return err
	}
	stage := filepath.Join(stages, id)
	if err = adminDirectory(stage); err != nil {
		return err
	}
	defer func() {
		if _, err := os.Stat(journalPath); errors.Is(err, os.ErrNotExist) {
			result = errors.Join(result, os.RemoveAll(stage))
		}
	}()
	if err = copyWindowsEngine(filepath.Join(filepath.Dir(self), "mihomo.exe"), filepath.Join(stage, "mihomo.exe")); err != nil {
		return err
	}
	s := &windowsSession{owner: owner, uplink: uplink, journalPath: journalPath, stage: stage, journal: windowsJournal{Version: 1, Device: "wg-" + id, Stage: id}}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		result = errors.Join(result, s.stop(cleanup))
	}()
	// Clear inherited engine switches/proxy variables. Forced IPv6 capture is
	// local to this helper and its child; no system environment is modified.
	systemRoot, err := windows.GetWindowsDirectory()
	if err != nil {
		return err
	}
	os.Clearenv()
	for k, v := range map[string]string{"SystemRoot": systemRoot, "WINDIR": systemRoot, "PATH": filepath.Join(systemRoot, "System32"), "TEMP": stage, "TMP": stage, "USERPROFILE": stage, "SKIP_SYSTEM_IPV6_CHECK": "1"} {
		if err = os.Setenv(k, v); err != nil {
			return err
		}
	}
	listener, err := winio.ListenPipe(pipeName(owner), &winio.PipeConfig{SecurityDescriptor: "D:P(D;;GA;;;NU)(A;;GA;;;SY)(A;;GA;;;BA)(A;;GRGW;;;" + owner.SID + ")", InputBufferSize: 8192, OutputBufferSize: 8192})
	if err != nil {
		return errors.New("tun: cannot create private helper pipe")
	}
	defer listener.Close()
	serveCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	s.idle = time.AfterFunc(45*time.Second, cancel)
	defer s.idle.Stop()
	// Closing the listener unblocks Accept on owner exit or idle timeout.
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-serveCtx.Done():
				listener.Close()
				return
			case <-ticker.C:
				alive, err := owner.Identity.Alive()
				if err != nil || !alive {
					cancel()
				}
			}
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		fd, ok := conn.(interface{ Fd() uintptr })
		var pid uint32
		if !ok || windows.GetNamedPipeClientProcessId(windows.Handle(fd.Fd()), &pid) != nil || int(pid) != owner.Identity.PID {
			conn.Close()
			continue
		}
		current, err := processOwner(int(pid))
		if err != nil || current.SID != owner.SID || current.Identity.Start != owner.Identity.Start {
			conn.Close()
			continue
		}
		attempted, err := s.serve(serveCtx, conn)
		conn.Close()
		if attempted {
			return err
		}
	}
}

func (s *windowsSession) serve(ctx context.Context, conn net.Conn) (bool, error) {
	closeOnCancel := context.AfterFunc(ctx, func() { conn.Close() })
	defer closeOnCancel()
	var sequence uint64
	session := ""
	for {
		conn.SetDeadline(time.Now().Add(Lease))
		var r Request
		if err := readMessage(conn, &r); err != nil {
			return s.started, err
		}
		response := Response{Version: ProtocolVersion, ID: r.ID}
		var err error
		if r.Version != ProtocolVersion || r.ID <= sequence {
			err = errors.New("tun: invalid protocol sequence")
		} else {
			sequence = r.ID
			switch r.Operation {
			case "status":
				if s.started {
					err = s.healthy()
					response.Device = s.journal.Device
				} else {
					_, err = windowsNetwork("")
				}
			case "cancel":
				if !s.started {
					json.NewEncoder(conn).Encode(response)
					return true, nil
				}
				err = errors.New("tun: active session requires stop")
			case "start":
				if s.started || r.Session != "" {
					err = errors.New("tun: session already attempted")
				} else {
					s.started = true
					if s.idle != nil {
						s.idle.Stop()
					}
					conn.SetDeadline(time.Now().Add(45 * time.Second))
					err = s.start(ctx, r)
					if err == nil {
						session, err = randomWindowsID()
						response.Session = session
						response.Device = s.journal.Device
					}
				}
			case "heartbeat", "stop":
				if session == "" || r.Session != session {
					err = errors.New("tun: session mismatch")
				} else if r.Operation == "stop" {
					cleanup, cancel := context.WithTimeout(context.Background(), 12*time.Second)
					err = s.stop(cleanup)
					cancel()
					if err != nil {
						response.Error = err.Error()
					}
					json.NewEncoder(conn).Encode(response)
					return true, err
				} else {
					err = s.healthy()
				}
			default:
				err = errors.New("tun: unsupported operation")
			}
		}
		if err != nil {
			response.Error = err.Error()
		}
		if encodeErr := json.NewEncoder(conn).Encode(response); encodeErr != nil {
			return s.started, encodeErr
		}
		if err != nil && s.started {
			return true, err
		}
	}
}

func (s *windowsSession) start(ctx context.Context, r Request) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.UDPKeepalive {
		return errors.New("tun: Windows TURN UDP transport is not supported yet; select TCP")
	}
	// Relay identity comes from the OS, not an executable path supplied by IPC.
	relay, err := processOwner(r.RelayPID)
	if err != nil || relay.SID != s.owner.SID || !strings.EqualFold(relay.Identity.Path, filepath.Join(filepath.Dir(s.owner.Identity.Path), "free-turn-proxy.exe")) {
		return errors.New("tun: unexpected relay process")
	}
	link, err := windowsNetwork("")
	if err != nil {
		return err
	}
	if !sameWindowsNetwork(link, s.uplink) {
		return errors.New("tun: uplink changed before activation")
	}
	peers, err := windowsTCPPeers(r.RelayPID)
	if err != nil {
		return err
	}
	config, err := compileWindows(r.ProxyPort, relay.Identity.Path, link, s.journal.Device, r.DirectPeer, peers)
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(s.stage, "config.json"), config, 0600); err != nil {
		return err
	}
	if err = writeWindowsJournal(s.journalPath, s.journal); err != nil {
		return err
	}
	s.process, err = (platform.Runner{}).Start(ctx, ports.Command{Path: filepath.Join(s.stage, "mihomo.exe"), Args: []string{"-d", s.stage, "-f", filepath.Join(s.stage, "config.json")}, Dir: s.stage, ClearProxyEnv: true})
	if err != nil {
		return err
	}
	deadline := time.NewTimer(35 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.process.Done():
			return errors.New("tun: privileged mihomo exited during startup")
		case <-deadline.C:
			return errors.New("tun: adapter/DNS readiness timed out")
		case <-tick.C:
			links, err := windowsLinks()
			if err != nil {
				return err
			}
			for _, l := range links {
				if l.Name == s.journal.Device && l.Up && len(l.Addresses) >= 2 && containsString(l.DNS, "198.18.0.2") && containsString(l.DNS, "fdfe:dcba:9876::2") {
					if err := windowsCaptureReady(l.Index); err == nil && windowsDNSReady() == nil {
						return nil
					}
				}
			}
		}
	}
}

func containsString(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func (s *windowsSession) healthy() error {
	if s.process == nil {
		return errors.New("tun: engine unavailable")
	}
	select {
	case <-s.process.Done():
		return errors.New("tun: privileged engine stopped")
	default:
	}
	link, err := windowsNetwork(s.journal.Device)
	if err != nil {
		return err
	}
	if !sameWindowsNetwork(link, s.uplink) {
		return errors.New("tun: uplink changed; reconnect required")
	}
	links, err := windowsLinks()
	if err != nil {
		return err
	}
	ready := false
	for _, l := range links {
		if l.Name == s.journal.Device && l.Up {
			if err = windowsCaptureReady(l.Index); err != nil {
				return err
			}
			ready = true
		}
	}
	if !ready {
		return errors.New("tun: adapter disappeared")
	}
	alive, err := s.owner.Identity.Alive()
	if err != nil || !alive {
		return errors.New("tun: owner disappeared")
	}
	return nil
}

func (s *windowsSession) stop(ctx context.Context) error {
	if s.cleaned {
		return nil
	}
	if s.process != nil {
		if err := s.process.Stop(ctx); err != nil {
			return err
		}
		s.process = nil
	}
	if err := recoverWindowsJournal(ctx, s.journalPath); err != nil {
		return err
	}
	s.cleaned = true
	return nil
}

func writeWindowsJournal(path string, j windowsJournal) error {
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return writeAdminFile(path, data)
}

func recoverWindowsJournal(ctx context.Context, path string) error {
	f, err := openComponent(path)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = requireAdminACL(windows.Handle(f.Fd())); err != nil {
		f.Close()
		return err
	}
	var j windowsJournal
	d := json.NewDecoder(io.LimitReader(f, 4097))
	d.DisallowUnknownFields()
	err = d.Decode(&j)
	if err == nil && d.Decode(new(any)) != io.EOF {
		err = errors.New("tun: trailing journal data")
	}
	f.Close()
	if err != nil || j.Version != 1 || !windowsDevicePattern.MatchString(j.Device) || j.Device != "wg-"+j.Stage {
		return errors.New("tun: invalid recovery journal; retained")
	}
	// Wintun-created adapters and their routes/DNS are handle-owned. WFP uses a
	// dynamic session. Both must disappear when Job Object closes the engine.
	// Never reset a physical adapter or delete an adapter merely by its name.
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		links, err := windowsLinks()
		if err != nil {
			return err
		}
		present := false
		for _, l := range links {
			if l.Name == j.Device {
				present = true
			}
		}
		if !present {
			break
		}
		select {
		case <-ctx.Done():
			return errors.New("tun: owned adapter still exists; recovery journal retained")
		case <-tick.C:
		}
	}
	stages, err := stageRoot()
	if err != nil {
		return err
	}
	stage := filepath.Join(stages, j.Stage)
	if _, err = os.Stat(stage); err == nil {
		if err = adminDirectory(stage); err != nil {
			return err
		}
		if err = os.RemoveAll(stage); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err = os.Remove(path); err != nil {
		return fmt.Errorf("tun: could not clear completed journal: %w", err)
	}
	return nil
}
