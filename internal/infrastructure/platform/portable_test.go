package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPortablePaths(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "portable.json")
	for _, location := range []string{"controller", filepath.Join("bin", "controller")} {
		executable := filepath.Join(root, location)
		if _, portable, err := portablePaths(executable); err != nil || portable {
			t.Fatalf("without marker: portable=%v err=%v", portable, err)
		}
	}
	// Empty marker is sufficient; neither JSON parsing nor OS config is required.
	if err := os.WriteFile(marker, nil, 0600); err != nil {
		t.Fatal(err)
	}
	for _, location := range []string{"controller", filepath.Join("bin", "controller")} {
		paths, portable, err := portablePaths(filepath.Join(root, location))
		if err != nil || !portable {
			t.Fatalf("portable=%v err=%v", portable, err)
		}
		if paths.Data != filepath.Join(root, "data") || paths.Runtime != filepath.Join(root, "data", "runtime") {
			t.Fatalf("unexpected paths: %+v", paths)
		}
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(marker, 0700); err != nil {
		t.Fatal(err)
	}
	if _, _, err := portablePaths(filepath.Join(root, "bin", "controller")); err == nil {
		t.Fatal("directory marker accepted")
	}
}
