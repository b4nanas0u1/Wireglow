package persistence

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

// Exercise the real Windows configuration directory, not only TEMP: the
// reported import succeeded there but reading the saved profile failed.
func TestProfileReadInWindowsConfigDirectory(t *testing.T) {
	config, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(config, "wireglow-read-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	store := &Profiles{Dir: filepath.Join(dir, "profiles")}
	link := "freeturn://" + base64.RawURLEncoding.EncodeToString([]byte(`{"v":1,"provider":"direct","peer":"relay.example:56000","name":"Config directory test"}`))
	imported, err := store.Import(context.Background(), link)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), imported.ID); err != nil {
		t.Fatal(err)
	}
	items, err := store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != imported.ID {
		t.Fatal("imported profile missing from refreshed list")
	}
}
