package persistence

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestAtomicReplaceWhileReaderOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lease.json")
	type lease struct {
		Sequence int `json:"sequence"`
	}
	if err := WriteJSON(path, lease{1}); err != nil {
		t.Fatal(err)
	}
	reader, err := openReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if err := WriteJSON(path, lease{2}); err != nil {
		t.Fatalf("reader blocked atomic publication: %v", err)
	}
	var old lease
	if err := json.NewDecoder(reader).Decode(&old); err != nil {
		t.Fatal(err)
	}
	if old.Sequence != 1 {
		t.Fatalf("existing reader lost its snapshot: %v", old)
	}
	var current lease
	if err := ReadJSON(path, &current); err != nil {
		t.Fatal(err)
	}
	if current.Sequence != 2 {
		t.Fatalf("new reader saw stale lease: %v", current)
	}
}
