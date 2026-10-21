package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func desktopRunDir(t *testing.T, runtimeDir string) string {
	t.Helper()
	entries, err := os.ReadDir(runtimeDir)
	if err != nil {
		t.Fatal(err)
	}
	var run string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "owner-") {
			if run != "" {
				t.Fatal("multiple owner directories")
			}
			run = filepath.Join(runtimeDir, entry.Name())
		}
	}
	if run == "" {
		t.Fatal("missing owner directory")
	}
	return run
}
