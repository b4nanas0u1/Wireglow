package persistence

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type SessionManifest struct {
	Version   int    `json:"version"`
	Directory string `json:"directory"`
}

// CleanupSession вызывается только под controller lock после завершения прежней группы.
func CleanupSession(runtime string) error {
	path := filepath.Join(runtime, "active-session.json")
	var manifest SessionManifest
	if err := ReadJSON(path, &manifest); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	name := manifest.Directory
	if manifest.Version != 1 || filepath.Base(name) != name || !strings.HasPrefix(name, "session-") || len(name) <= len("session-") {
		return errors.New("storage: invalid session manifest")
	}
	if err := os.RemoveAll(filepath.Join(runtime, name)); err != nil {
		return err
	}
	return os.Remove(path)
}
