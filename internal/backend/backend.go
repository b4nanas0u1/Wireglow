// Package backend собирает Go API из application и адаптеров; RPC transport сюда не входит.
package backend

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	"wireglow/internal/application/controller"
	"wireglow/internal/application/ports"
	"wireglow/internal/domain/proxymode"
	"wireglow/internal/infrastructure/engines"
	"wireglow/internal/infrastructure/network"
	"wireglow/internal/infrastructure/persistence"
	"wireglow/internal/infrastructure/platform"
	"wireglow/internal/infrastructure/tun"
)

type Options struct {
	Paths            platform.Paths
	Relay            engines.Binary
	Mihomo           engines.Binary
	ProbeURL         string
	ProbeStatus      int
	SystemProxy      bool
	RecoveryIdentity platform.Identity
}

type Backend struct {
	Controller   *controller.Controller
	Profiles     *persistence.Profiles
	lock         *platform.Lock
	settingsPath string
	settingsMu   sync.Mutex
	tun          *tun.Client
	once         sync.Once
	err          error
}

type manualNetwork struct{ manager *network.Manager }

type guardedNetwork struct {
	manager  *network.Manager
	identity platform.Identity
}

func (n guardedNetwork) Recover(ctx context.Context) error { return n.manager.Recover(ctx) }
func (n guardedNetwork) Apply(ctx context.Context, address string) error {
	if n.identity.PID == os.Getpid() {
		return errors.New("network: recovery must run outside controller")
	}
	alive, err := n.identity.Alive()
	if err != nil || !alive {
		return errors.New("network: independent recovery not ready")
	}
	return n.manager.Apply(ctx, address)
}

func (n manualNetwork) Recover(ctx context.Context) error { return n.manager.Recover(ctx) }
func (manualNetwork) Apply(context.Context, string) error { return nil }

func Open(ctx context.Context, o Options) (*Backend, error) {
	if err := platform.EnsurePaths(o.Paths); err != nil {
		return nil, err
	}
	lock, err := platform.AcquireLock(filepath.Join(o.Paths.Runtime, "controller.lock"))
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			_ = lock.Close()
		}
	}()
	manager := &network.Manager{Path: filepath.Join(o.Paths.Data, "proxy-journal.json"), Settings: platform.ProxySettings{}}
	if err := errors.Join(manager.Recover(ctx), persistence.CleanupSession(o.Paths.Runtime)); err != nil {
		return nil, err
	}
	preparer := &engines.Preparer{Relay: o.Relay, Mihomo: o.Mihomo, RuntimeDir: o.Paths.Runtime, ProbeURL: o.ProbeURL, ProbeStatus: o.ProbeStatus}
	local := manualNetwork{manager}
	defaultMode := proxymode.Local
	if o.SystemProxy {
		defaultMode = proxymode.System
	} // Только для отдельной CLI-команды connect.
	tunClient := &tun.Client{}
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	if err = tunClient.ConfigurePortable(executable); err != nil {
		return nil, err
	}
	options := controller.Options{DefaultMode: defaultMode, NetworkForMode: func(mode proxymode.Mode) (ports.Network, error) {
		if mode == proxymode.Tun {
			if runtime.GOOS != "linux" {
				return nil, errors.New("tun: Windows helper not implemented")
			}
			return tunClient, nil
		}
		if mode == proxymode.Local {
			return local, nil
		}
		return guardedNetwork{manager: manager, identity: o.RecoveryIdentity}, nil
	}}
	b := &Backend{tun: tunClient, Controller: controller.New(platform.Runner{}, preparer, local, options), Profiles: &persistence.Profiles{Dir: filepath.Join(o.Paths.Data, "profiles")}, lock: lock, settingsPath: filepath.Join(o.Paths.Data, "client-settings.json")}
	tunClient.RelayPID = b.Controller.RelayPID
	preparer.Addresses = func() (string, string, string, error) {
		s, err := b.GetSettings()
		return s.RelayAddress, s.ProxyAddress, net.JoinHostPort("127.0.0.1", strconv.Itoa(s.APIPort)), err
	}
	ok = true
	return b, nil
}

func (b *Backend) Connect(ctx context.Context, id string) error {
	p, err := b.Profiles.Get(ctx, id)
	if err != nil {
		return err
	}
	return b.Controller.Connect(ctx, p)
}

func (b *Backend) Close() error {
	b.once.Do(func() { b.err = errors.Join(b.Controller.Shutdown(), b.lock.Close()) })
	return b.err
}

func (b *Backend) TunUnavailable(ctx context.Context) string {
	if b.tun == nil {
		return "TUN helper unavailable"
	}
	return b.tun.Unavailable(ctx)
}
