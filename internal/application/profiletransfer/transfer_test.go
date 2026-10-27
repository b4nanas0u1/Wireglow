package profiletransfer_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"wireglow/internal/application/profiletransfer"
	"wireglow/internal/domain/profile"
	"wireglow/internal/freeturn"
	"wireglow/internal/infrastructure/persistence"
)

func node(id, name string) profile.Profile {
	return profile.Profile{ID: strings.Repeat(id, 32), Connection: freeturn.Connection{Name: name, Provider: "vk", Peer: "relay.example:443", VK: &freeturn.VKSettings{CallLinks: []string{"https://vk.ru/call/join/synthetic"}, TurnAddress: "[2001:db8::1]:443", Platform: "mobile", ManualCaptcha: true}, DNS: freeturn.DNSSettings{Mode: "doh", Servers: []string{"1.1.1.1"}}, WG: &freeturn.WireGuardSettings{Config: "synthetic private config"}, Obfuscation: freeturn.Obfuscation{Mode: "none"}}}
}
func fixture(t *testing.T) (*profiletransfer.Service, string) {
	t.Helper()
	dir := t.TempDir()
	store := &persistence.Profiles{Dir: filepath.Join(dir, "profiles")}
	return &profiletransfer.Service{Store: store, Files: persistence.ProfileFiles{ProfilesDir: store.Dir}}, filepath.Join(dir, "backup.json")
}
func archive(t *testing.T, path string, items ...profile.Profile) {
	t.Helper()
	data, err := json.Marshal(profiletransfer.Archive{Format: "wireglow-profiles", Version: 1, Profiles: items})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}
func TestRoundTripAndConflictPolicies(t *testing.T) {
	ctx := context.Background()
	s, path := fixture(t)
	p := node("a", "VK Calls")
	archive(t, path, p)
	preview, err := s.Preview(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Entries[0].Existing != nil {
		t.Fatal("unexpected conflict")
	}
	// Preview is a snapshot: a changed source file must not replace the reviewed data.
	archive(t, path, node("b", "other"))
	result, err := s.Import(ctx, preview.Token, []profiletransfer.Choice{{Index: 0, Action: "add"}})
	if err != nil || result[0].Status != "imported" {
		t.Fatal(result, err)
	}
	got, err := s.Store.Get(ctx, p.ID)
	if err != nil || !reflect.DeepEqual(got, p) {
		t.Fatal("round trip lost profile settings", err)
	}
	if _, err = s.Import(ctx, preview.Token, []profiletransfer.Choice{{Index: 0, Action: "copy"}}); !errors.Is(err, profiletransfer.ErrExpired) {
		t.Fatal("preview was replayable", err)
	}
	if err = s.Export(ctx, []string{p.ID}, path); err != nil {
		t.Fatal(err)
	}
	bytes, _ := os.ReadFile(path)
	decoded, err := profiletransfer.Decode(bytes)
	if err != nil || !reflect.DeepEqual(decoded.Profiles[0], p) {
		t.Fatal("export lost data", err)
	}
	if stat, _ := os.Stat(path); runtime.GOOS != "windows" && stat.Mode().Perm() != 0600 {
		t.Fatal("export is not private")
	}
	for _, action := range []string{"skip", "copy", "replace"} {
		changed := p
		changed.Connection.Name = "Imported name"
		archive(t, path, changed)
		preview, err = s.Preview(ctx, path)
		if err != nil || preview.Entries[0].Existing == nil {
			t.Fatal("conflict not detected", err)
		}
		result, err = s.Import(ctx, preview.Token, []profiletransfer.Choice{{Index: 0, Action: action}})
		if err != nil {
			t.Fatal(err)
		}
		got, _ = s.Store.Get(ctx, p.ID)
		if action == "replace" && got.Connection.Name != "Imported name" {
			t.Fatal("replace ignored")
		}
		if action != "replace" && !reflect.DeepEqual(got, p) {
			t.Fatal("existing profile changed")
		}
		if action == "skip" && result[0].Status != "skipped" {
			t.Fatal(result)
		}
	}
	list, _ := s.Store.List(ctx)
	if len(list) != 2 {
		t.Fatal("copy did not get a fresh ID", list)
	}
}
func TestRejectsUnsafeAndMalformedArchives(t *testing.T) {
	p := node("a", "test")
	good, _ := json.Marshal(profiletransfer.Archive{Format: "wireglow-profiles", Version: 1, Profiles: []profile.Profile{p}})
	for _, data := range [][]byte{nil, []byte(`{}`), append(append([]byte{}, good...), []byte(`{}`)...), []byte(strings.Replace(string(good), `"version":1`, `"version":2`, 1)), []byte(strings.Replace(string(good), `"format":`, `"unknown":"secret","format":`, 1)), []byte(strings.Repeat("x", profiletransfer.MaxBytes+1))} {
		if _, err := profiletransfer.Decode(data); !errors.Is(err, profiletransfer.ErrInvalid) {
			t.Fatalf("malformed input accepted, size %d", len(data))
		}
	}
	for _, items := range [][]profile.Profile{{p, p}, {node("z", "bad ID")}, {node("a", strings.Repeat("x", 81))}} {
		data, _ := json.Marshal(profiletransfer.Archive{Format: "wireglow-profiles", Version: 1, Profiles: items})
		if _, err := profiletransfer.Decode(data); err == nil {
			t.Fatal("invalid profiles accepted")
		}
	}
}
func TestStalePreviewAndPartialFailures(t *testing.T) {
	ctx := context.Background()
	s, path := fixture(t)
	p := node("a", "initial")
	archive(t, path, p)
	preview, _ := s.Preview(ctx, path)
	_, err := s.Import(ctx, preview.Token, []profiletransfer.Choice{{Index: 0, Action: "add"}})
	if err != nil {
		t.Fatal(err)
	}
	archive(t, path, p, node("b", "new"))
	preview, err = s.Preview(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	store := s.Store.(*persistence.Profiles)
	p.Connection.Name = "edited after preview"
	if err = store.Update(ctx, p); err != nil {
		t.Fatal(err)
	}
	results, err := s.Import(ctx, preview.Token, []profiletransfer.Choice{{Index: 0, Action: "replace"}, {Index: 1, Action: "add"}})
	if err != nil || results[0].Status != "changed" || results[1].Status != "imported" {
		t.Fatal(results, err)
	}
	saved, _ := store.Get(ctx, p.ID)
	if saved.Connection.Name != p.Connection.Name {
		t.Fatal("overwrote concurrent edit")
	}
}
func TestSelectionsValidatedBeforeWritesAndExportCannotOverwriteStore(t *testing.T) {
	ctx := context.Background()
	s, path := fixture(t)
	p := node("a", "test")
	archive(t, path, p)
	preview, _ := s.Preview(ctx, path)
	if _, err := s.Import(ctx, preview.Token, []profiletransfer.Choice{{Index: 0, Action: "add"}, {Index: 0, Action: "add"}}); !errors.Is(err, profiletransfer.ErrInvalid) {
		t.Fatal(err)
	}
	list, _ := s.Store.List(ctx)
	if len(list) != 0 {
		t.Fatal("invalid selection wrote profiles")
	}
	if _, err := s.Import(ctx, preview.Token, []profiletransfer.Choice{{Index: 0, Action: "replace"}}); !errors.Is(err, profiletransfer.ErrInvalid) {
		t.Fatal(err)
	}
	_, err := s.Import(ctx, preview.Token, []profiletransfer.Choice{{Index: 0, Action: "add"}})
	if err != nil {
		t.Fatal(err)
	}
	store := s.Store.(*persistence.Profiles)
	if err = s.Export(ctx, []string{p.ID}, filepath.Join(store.Dir, p.ID+".json")); !errors.Is(err, profiletransfer.ErrInvalid) {
		t.Fatal("overwrote store", err)
	}
	saved, _ := s.Store.Get(ctx, p.ID)
	if !reflect.DeepEqual(saved, p) {
		t.Fatal("profile store corrupted")
	}
}
func TestFileReadAndWriteErrorsPreserveExistingData(t *testing.T) {
	ctx := context.Background()
	s, path := fixture(t)
	archive(t, path, node("a", "test"))
	preview, _ := s.Preview(ctx, path)
	// A missing parent produces a genuine storage error without requiring chmod semantics.
	store := s.Store.(*persistence.Profiles)
	if err := os.WriteFile(store.Dir, []byte("block directory"), 0600); err != nil {
		t.Fatal(err)
	}
	results, err := s.Import(ctx, preview.Token, []profiletransfer.Choice{{Index: 0, Action: "add"}})
	if err != nil || results[0].Status != "failed" {
		t.Fatal(results, err)
	}
	if _, err = s.Preview(ctx, filepath.Dir(path)); err == nil {
		t.Fatal("directory accepted as a file")
	}
}
