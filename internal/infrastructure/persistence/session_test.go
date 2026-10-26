package persistence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionRecoveryRemovesOnlyRecordedPrivateDirectory(t *testing.T) {
	runtime := t.TempDir()
	session := filepath.Join(runtime, "session-one")
	other := filepath.Join(runtime, "unrelated")
	if err := os.Mkdir(session, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(other, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(session, "config.yaml"), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(filepath.Join(runtime, "active-session.json"), SessionManifest{1, "session-one"}); err != nil {
		t.Fatal(err)
	}
	if err := CleanupSession(runtime); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(session); !os.IsNotExist(err) {
		t.Fatal("secret session survived")
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatal("unrelated directory deleted")
	}
	if err := CleanupSession(runtime); err != nil {
		t.Fatal(err)
	}
}

func TestSessionRecoveryRejectsTraversal(t *testing.T) {
	runtime := t.TempDir()
	for _, name := range []string{"../session-other", "/tmp/session-other", "unrelated", "session-"} {
		if err := WriteJSON(filepath.Join(runtime, "active-session.json"), SessionManifest{1, name}); err != nil {
			t.Fatal(err)
		}
		if err := CleanupSession(runtime); err == nil {
			t.Fatal("unsafe manifest accepted", name)
		}
	}
}
