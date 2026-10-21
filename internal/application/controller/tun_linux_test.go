package controller

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"wireglow/internal/application/ports"
	"wireglow/internal/domain/profile"
	"wireglow/internal/domain/proxymode"
)

type healthyTun struct {
	modeNetwork
	failed chan struct{}
}

func (n *healthyTun) Failed() <-chan struct{} { return n.failed }
func TestTunOwnerFailureStopsSession(t *testing.T) {
	c, r, _, plan, _ := harness()
	defer c.Shutdown()
	plan.TunReady = true
	tun := &healthyTun{failed: make(chan struct{})}
	c.options.NetworkForMode = func(proxymode.Mode) (ports.Network, error) { return tun, nil }
	if err := c.ConnectWithMode(context.Background(), profile.Profile{}, proxymode.Tun); err != nil {
		t.Fatal(err)
	}
	if c.Snapshot().ActiveMode != proxymode.Tun {
		t.Fatal(c.Snapshot())
	}
	close(tun.failed)
	deadline := time.Now().Add(time.Second)
	for c.Snapshot().State != Failed && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if c.Snapshot().State != Failed {
		t.Fatal("helper loss left active session")
	}
	for _, p := range r.processes {
		select {
		case <-p.Done():
		default:
			t.Fatal("process not stopped")
		}
	}
}
func TestTunSwitchRequiresBootstrapWithoutTouchingActiveSession(t *testing.T) {
	c, r, _, _, tr := switchHarness(t)
	defer c.Shutdown()
	before := c.Snapshot()
	if err := c.SwitchMode(context.Background(), proxymode.Tun); err == nil {
		t.Fatal("unprepared relay accepted")
	}
	assertCoresUnchanged(t, c, r, before, tr)
	if c.Snapshot().ActiveMode != proxymode.Local {
		t.Fatal(c.Snapshot())
	}
}

type diagnosticTun struct{ modeNetwork }

func (*diagnosticTun) Apply(context.Context, string) error { return errors.New("private detail") }
func (*diagnosticTun) NetworkDiagnostic() string {
	return "tun: temporary executable memory: permission denied"
}
func TestTunActivationPublishesAdapterDiagnostic(t *testing.T) {
	c, _, _, plan, _ := harness()
	defer c.Shutdown()
	plan.TunReady = true
	c.options.NetworkForMode = func(proxymode.Mode) (ports.Network, error) { return &diagnosticTun{}, nil }
	if err := c.ConnectWithMode(context.Background(), profile.Profile{}, proxymode.Tun); err == nil {
		t.Fatal("failure expected")
	}
	message := c.Snapshot().Error
	if !strings.Contains(message, "temporary executable memory: permission denied") || strings.Contains(message, "private detail") {
		t.Fatal("unsafe or missing diagnostic", message)
	}
}

type safeProbeError string

func (e safeProbeError) Error() string            { return "private engine error" }
func (e safeProbeError) PublicDiagnostic() string { return string(e) }
func TestPostActivationProbePublishesOnlySanitizedCause(t *testing.T) {
	for _, safe := range []bool{false, true} {
		c, _, _, plan, _ := harness()
		plan.TunReady = true
		checks := 0
		plan.Probe = func(context.Context) error {
			checks++
			if checks == 1 {
				return nil
			}
			if safe {
				return safeProbeError("probe: UDP DNS timeout after 10s")
			}
			return errors.New("private engine error")
		}
		if err := c.ConnectWithMode(context.Background(), profile.Profile{}, proxymode.Tun); err == nil {
			t.Fatal("failure expected")
		}
		message := c.Snapshot().Error
		if strings.Contains(message, "private engine error") || strings.Contains(message, "UDP DNS timeout") != safe {
			t.Fatal(message)
		}
		if !strings.HasPrefix(message, "probe after activation failed") {
			t.Fatal(message)
		}
		c.Shutdown()
	}
}
