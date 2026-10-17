package controller

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wireglow/internal/application/ports"
	"wireglow/internal/domain/profile"
	"wireglow/internal/domain/proxymode"
	"wireglow/internal/infrastructure/network"
)

type proxySettings struct {
	values map[string]string
	write  func(string, string) error
}

func (s *proxySettings) Desired(address string) ([]network.Change, error) {
	return []network.Change{{Key: "host", After: address}, {Key: "mode", After: "manual"}}, nil
}
func (s *proxySettings) Read(ctx context.Context, key string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return s.values[key], nil
}
func (s *proxySettings) Write(ctx context.Context, key, value string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if s.write != nil {
		if err := s.write(key, value); err != nil {
			return err
		}
	}
	s.values[key] = value
	return nil
}

type localProxy struct{ manager *network.Manager }

func (n localProxy) Recover(ctx context.Context) error { return n.manager.Recover(ctx) }
func (localProxy) Apply(context.Context, string) error { return nil }

func switchHarness(t *testing.T) (*Controller, *fakeRunner, *proxySettings, *network.Manager, *trace) {
	t.Helper()
	c, r, _, _, tr := harness()
	settings := &proxySettings{values: map[string]string{"host": "previous.proxy", "mode": "auto"}}
	manager := &network.Manager{Path: filepath.Join(t.TempDir(), "proxy-journal.json"), Settings: settings}
	local := localProxy{manager}
	c.network = local
	c.options.NetworkForMode = func(mode proxymode.Mode) (ports.Network, error) {
		if mode == proxymode.Local {
			return local, nil
		}
		return manager, nil
	}
	t.Cleanup(func() { settings.write = nil; _ = c.Shutdown() })
	if err := c.Connect(context.Background(), profile.Profile{ID: "one"}); err != nil {
		t.Fatal(err)
	}
	return c, r, settings, manager, tr
}

func assertCoresUnchanged(t *testing.T, c *Controller, r *fakeRunner, before Snapshot, tr *trace) {
	t.Helper()
	after := c.Snapshot()
	if after.State != Connected || after.Generation != before.Generation || after.ProfileID != before.ProfileID || after.MixedAddress != before.MixedAddress || after.Latency != before.Latency {
		t.Fatalf("session changed: before=%+v after=%+v", before, after)
	}
	if len(r.processes) != 2 || strings.Contains(tr.text(), "stop:") {
		t.Fatal("cores restarted", tr.text())
	}
	for _, p := range r.processes {
		select {
		case <-p.Done():
			t.Fatal("core stopped")
		default:
		}
	}
}

func TestSwitchModeChangesOnlyOSSettings(t *testing.T) {
	c, r, s, m, tr := switchHarness(t)
	before := c.Snapshot()
	for i := 0; i < 2; i++ {
		for _, mode := range []proxymode.Mode{proxymode.System, proxymode.System, proxymode.Local} {
			if err := c.SwitchMode(context.Background(), mode); err != nil {
				t.Fatal(err)
			}
			assertCoresUnchanged(t, c, r, before, tr)
			if c.Snapshot().ActiveMode != mode {
				t.Fatal(c.Snapshot())
			}
			if mode == proxymode.System {
				if s.values["host"] != before.MixedAddress || s.values["mode"] != "manual" {
					t.Fatal(s.values)
				}
			} else {
				if s.values["host"] != "previous.proxy" || s.values["mode"] != "auto" {
					t.Fatal(s.values)
				}
				if _, err := os.Stat(m.Path); !os.IsNotExist(err) {
					t.Fatalf("journal survived: %v", err)
				}
			}
		}
	}
}

func TestSwitchModeRollsBackPartialApplyWithoutStoppingCores(t *testing.T) {
	c, r, s, m, tr := switchHarness(t)
	before := c.Snapshot()
	s.write = func(key, value string) error {
		if key == "mode" && value == "manual" {
			return errors.New("write failed")
		}
		return nil
	}
	if err := c.SwitchMode(context.Background(), proxymode.System); !errors.Is(err, ErrModeSwitch) {
		t.Fatal(err)
	}
	assertCoresUnchanged(t, c, r, before, tr)
	if c.Snapshot().ActiveMode != proxymode.Local || c.Snapshot().Error == "" || s.values["host"] != "previous.proxy" || s.values["mode"] != "auto" {
		t.Fatal(c.Snapshot(), s.values)
	}
	if _, err := os.Stat(m.Path); !os.IsNotExist(err) {
		t.Fatal("rollback left journal", err)
	}
	s.write = nil
	if err := c.SwitchMode(context.Background(), proxymode.System); err != nil {
		t.Fatal(err)
	}
	if c.Snapshot().Error != "" {
		t.Fatal(c.Snapshot())
	}
	assertCoresUnchanged(t, c, r, before, tr)
}

func TestSwitchModeRecoveryFailureRetainsJournalAndCores(t *testing.T) {
	c, r, s, m, tr := switchHarness(t)
	before := c.Snapshot()
	if err := c.SwitchMode(context.Background(), proxymode.System); err != nil {
		t.Fatal(err)
	}
	s.write = func(key, value string) error {
		if value == "auto" {
			return errors.New("restore failed")
		}
		return nil
	}
	if err := c.SwitchMode(context.Background(), proxymode.Local); !errors.Is(err, ErrModeSwitch) {
		t.Fatal(err)
	}
	assertCoresUnchanged(t, c, r, before, tr)
	if c.Snapshot().ActiveMode != "" || c.Snapshot().Error == "" {
		t.Fatal("claimed known mode after partial recovery", c.Snapshot())
	}
	if _, err := os.Stat(m.Path); err != nil {
		t.Fatal("lost recovery journal", err)
	}
	s.write = nil
	if err := c.SwitchMode(context.Background(), proxymode.Local); err != nil {
		t.Fatal(err)
	}
	if c.Snapshot().ActiveMode != proxymode.Local || c.Snapshot().Error != "" || s.values["mode"] != "auto" {
		t.Fatal(c.Snapshot(), s.values)
	}
	assertCoresUnchanged(t, c, r, before, tr)
}

func TestSwitchModeCancellationStillRollsBack(t *testing.T) {
	c, r, s, _, tr := switchHarness(t)
	before := c.Snapshot()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.write = func(key, value string) error {
		if key == "host" && value == before.MixedAddress {
			cancel()
		}
		return nil
	}
	if err := c.SwitchMode(ctx, proxymode.System); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	assertCoresUnchanged(t, c, r, before, tr)
	if s.values["host"] != "previous.proxy" || c.Snapshot().ActiveMode != proxymode.Local {
		t.Fatal(s.values, c.Snapshot())
	}
}

func TestSwitchModeRequiresActiveSession(t *testing.T) {
	c, _, _, _, _ := switchHarness(t)
	if err := c.SwitchMode(context.Background(), "tun"); !errors.Is(err, ErrModeSwitch) {
		t.Fatal(err)
	}
	if err := c.Disconnect(); err != nil {
		t.Fatal(err)
	}
	if err := c.SwitchMode(context.Background(), proxymode.System); !errors.Is(err, ErrBusy) {
		t.Fatal(err)
	}
}

func TestSwitchModeKeepsJournalWhenApplyAndRollbackFail(t *testing.T) {
	c, r, s, m, tr := switchHarness(t)
	before := c.Snapshot()
	s.write = func(key, value string) error {
		if key == "mode" && value == "manual" || key == "host" && value == "previous.proxy" {
			return errors.New("write blocked")
		}
		return nil
	}
	if err := c.SwitchMode(context.Background(), proxymode.System); !errors.Is(err, ErrModeSwitch) {
		t.Fatal(err)
	}
	assertCoresUnchanged(t, c, r, before, tr)
	if c.Snapshot().ActiveMode != "" || c.Snapshot().Error == "" {
		t.Fatal(c.Snapshot())
	}
	if _, err := os.Stat(m.Path); err != nil {
		t.Fatal("lost rollback journal", err)
	}
	s.write = nil
	if err := c.SwitchMode(context.Background(), proxymode.Local); err != nil {
		t.Fatal(err)
	}
	if s.values["host"] != "previous.proxy" || c.Snapshot().ActiveMode != proxymode.Local {
		t.Fatal(s.values, c.Snapshot())
	}
	assertCoresUnchanged(t, c, r, before, tr)
}
