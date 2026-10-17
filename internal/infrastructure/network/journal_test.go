package network

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"wireglow/internal/infrastructure/persistence"
)

type fakeSettings struct {
	values map[string]string
	write  func(string, string) error
	writes []string
}

func (s *fakeSettings) Desired(string) ([]Change, error) {
	return []Change{{Key: "host", After: "local"}, {Key: "mode", After: "manual"}}, nil
}
func (s *fakeSettings) Read(ctx context.Context, key string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return s.values[key], nil
}
func (s *fakeSettings) Write(ctx context.Context, key, value string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	s.writes = append(s.writes, key+"="+value)
	if s.write != nil {
		return s.write(key, value)
	}
	s.values[key] = value
	return nil
}
func journalHarness(t *testing.T) (*Manager, *fakeSettings) {
	t.Helper()
	s := &fakeSettings{values: map[string]string{"host": "old", "mode": "auto"}}
	return &Manager{Path: filepath.Join(t.TempDir(), "journal.json"), Settings: s}, s
}

func TestJournalWrittenBeforeEachMutation(t *testing.T) {
	m, s := journalHarness(t)
	s.write = func(key, value string) error {
		var j Journal
		if err := persistence.ReadJSON(m.Path, &j); err != nil {
			t.Fatal(err)
		}
		last := j.Changes[len(j.Changes)-1]
		if last.Key != key || last.After != value || last.Before != s.values[key] {
			t.Fatal("missing durable intent")
		}
		s.values[key] = value
		return nil
	}
	if err := m.Apply(context.Background(), "ignored"); err != nil {
		t.Fatal(err)
	}
	s.write = nil
	if err := m.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.values["host"] != "old" || s.values["mode"] != "auto" {
		t.Fatal(s.values)
	}
	if s.writes[2] != "mode=auto" || s.writes[3] != "host=old" {
		t.Fatal("incorrect rollback order", s.writes)
	}
	if err := m.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestRecoverCrashBeforeOrAfterWrite(t *testing.T) {
	for _, after := range []bool{false, true} {
		t.Run(map[bool]string{false: "before", true: "after"}[after], func(t *testing.T) {
			m, s := journalHarness(t)
			s.write = func(key, value string) error {
				if after {
					s.values[key] = value
				}
				return errors.New("simulated crash")
			}
			if err := m.Apply(context.Background(), ""); err == nil {
				t.Fatal("expected crash")
			}
			s.write = nil
			// Новый объект, как при независимом восстановлении после аварии controller.
			other := &Manager{Path: m.Path, Settings: s}
			if err := other.Recover(context.Background()); err != nil {
				t.Fatal(err)
			}
			if s.values["host"] != "old" || s.values["mode"] != "auto" {
				t.Fatal(s.values)
			}
		})
	}
}

func TestRecoverPreservesExternalChanges(t *testing.T) {
	m, s := journalHarness(t)
	if err := m.Apply(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	s.values["host"] = "other-vpn"
	if err := m.Recover(context.Background()); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	if s.values["host"] != "other-vpn" || s.values["mode"] != "auto" {
		t.Fatal(s.values)
	}
	if _, err := os.Stat(m.Path); err != nil {
		t.Fatal("conflict journal lost")
	}
	if err := m.Apply(context.Background(), ""); !errors.Is(err, ErrConflict) {
		t.Fatal("new session overwrote conflict", err)
	}
}

func TestFailedRecoveryRetainsJournalForNextStartup(t *testing.T) {
	m, s := journalHarness(t)
	if err := m.Apply(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	s.write = func(string, string) error { return errors.New("permission denied") }
	if err := m.Recover(context.Background()); err == nil {
		t.Fatal("missing error")
	}
	s.write = nil
	if err := m.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.values["mode"] != "auto" || s.values["host"] != "old" {
		t.Fatal(s.values)
	}
}

func TestCorruptJournalDoesNotChangeSettings(t *testing.T) {
	m, s := journalHarness(t)
	if err := os.WriteFile(m.Path, []byte(`{"version":77,"changes":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Recover(context.Background()); err == nil {
		t.Fatal("unknown version accepted")
	}
	if len(s.writes) != 0 {
		t.Fatal("settings mutated")
	}
}
