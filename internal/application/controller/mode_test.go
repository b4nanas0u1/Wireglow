package controller

import (
	"context"
	"errors"
	"testing"

	"wireglow/internal/application/ports"
	"wireglow/internal/domain/profile"
	"wireglow/internal/domain/proxymode"
)

type modeNetwork struct{ applies, recoveries int }

func (n *modeNetwork) Apply(context.Context, string) error { n.applies++; return nil }
func (n *modeNetwork) Recover(context.Context) error       { n.recoveries++; return nil }

func TestSessionModeSwitchUsesCorrectNetworkAndCleanup(t *testing.T) {
	c, _, _, _, _ := harness()
	defer c.Shutdown()
	local, system := &modeNetwork{}, &modeNetwork{}
	selections := 0
	c.options.NetworkForMode = func(mode proxymode.Mode) (ports.Network, error) {
		selections++
		if mode == proxymode.System {
			return system, nil
		}
		return local, nil
	}
	ctx := context.Background()
	if err := c.ConnectWithMode(ctx, profile.Profile{}, "tun"); err == nil {
		t.Fatal("unsupported mode accepted")
	}
	if selections != 0 {
		t.Fatal("invalid mode selected network")
	}
	if err := c.ConnectWithMode(ctx, profile.Profile{}, proxymode.System); err != nil {
		t.Fatal(err)
	}
	if c.Snapshot().ActiveMode != proxymode.System || system.applies != 1 || local.applies != 0 {
		t.Fatal("wrong active mode/network")
	}
	if err := c.ConnectWithMode(ctx, profile.Profile{}, proxymode.Local); !errors.Is(err, ErrBusy) {
		t.Fatalf("active switch = %v", err)
	}
	if selections != 1 {
		t.Fatal("busy request changed network")
	}
	if err := c.Disconnect(); err != nil {
		t.Fatal(err)
	}
	if system.recoveries != 2 || c.Snapshot().ActiveMode != "" {
		t.Fatal("system session not cleaned up")
	}
	if err := c.ConnectWithMode(ctx, profile.Profile{}, proxymode.Local); err != nil {
		t.Fatal(err)
	}
	if system.applies != 1 || local.applies != 1 || c.Snapshot().ActiveMode != proxymode.Local {
		t.Fatal("next session did not use local mode")
	}
	if err := c.Disconnect(); err != nil {
		t.Fatal(err)
	}
	if local.recoveries != 2 {
		t.Fatal("local session did not clean up")
	}
}
