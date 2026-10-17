package persistence

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProfileRoundTripAndSummaryRedaction(t *testing.T) {
	store := &Profiles{Dir: t.TempDir()}
	link := "freeturn://" + base64.RawURLEncoding.EncodeToString([]byte(`{"v":1,"provider":"direct","peer":"relay.example:56000","key":"secret-value","name":"Test"}`))
	summary, err := store.Import(context.Background(), link)
	if err != nil {
		t.Fatal(err)
	}
	p, err := store.Get(context.Background(), summary.ID)
	if err != nil {
		t.Fatal(err)
	}
	if p.Connection.Obfuscation.Key != "secret-value" {
		t.Fatal("secrets not stored")
	}
	list, err := store.List(context.Background())
	if err != nil || len(list) != 1 {
		t.Fatal(list, err)
	}
	if list[0].Peer != "relay.example:56000" {
		t.Fatal("public peer missing")
	}
	data, _ := json.Marshal(list)
	if strings.Contains(string(data), "secret") {
		t.Fatal("summary exposes secrets")
	}
	stat, err := os.Stat(filepath.Join(store.Dir, summary.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && stat.Mode().Perm()&0077 != 0 {
		t.Fatal("profile readable by others")
	}
	if err := store.Delete(context.Background(), summary.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), summary.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}

func TestProfileIDsCannotEscapeDirectory(t *testing.T) {
	store := &Profiles{Dir: t.TempDir()}
	for _, id := range []string{"../secret", "/etc/passwd", "", "ABCDEF0123456789ABCDEF0123456789AB"} {
		if _, err := store.Get(context.Background(), id); !errors.Is(err, ErrNotFound) {
			t.Fatal(id, err)
		}
		if err := store.Delete(context.Background(), id); !errors.Is(err, ErrNotFound) {
			t.Fatal(id, err)
		}
	}
}

func TestWriteAtomicReplacesCompleteFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "value.json")
	if err := WriteJSON(path, map[string]int{"old": 1}); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(path, map[string]int{"new": 2}); err != nil {
		t.Fatal(err)
	}
	var value map[string]int
	if err := ReadJSON(path, &value); err != nil {
		t.Fatal(err)
	}
	if len(value) != 1 || value["new"] != 2 {
		t.Fatal(value)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatal("temporary file leaked")
	}
}
