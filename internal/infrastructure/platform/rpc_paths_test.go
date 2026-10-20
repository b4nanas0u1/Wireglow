package platform

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPrivateRPCSocketLifecycle(t *testing.T) {
	base, err := os.MkdirTemp("", "wg-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(base) })
	if runtime.GOOS == "windows" {
		t.Setenv("TMP", base)
		t.Setenv("TEMP", base)
	} else {
		t.Setenv("XDG_RUNTIME_DIR", base)
	}
	runID := "0123456789abcdef0123456789abcdef"
	path, err := RPCSocketPath(runID)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(filepath.Dir(path)) != base {
		t.Fatal("socket escaped system runtime")
	}
	cleanup, err := PrepareRPCSocket(runID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cleanup() })
	// Never reuse or overwrite an existing IPC directory.
	if _, err := PrepareRPCSocket(runID); !os.IsExist(err) {
		t.Fatalf("existing IPC directory accepted: %v", err)
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Dir(path))
		if err != nil || info.Mode().Perm() != 0700 {
			t.Fatalf("IPC directory is not private: %v", err)
		}
	}
	listener.Close()
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatalf("IPC directory survived cleanup: %v", err)
	}
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup is not idempotent: %v", err)
	}
	for _, invalid := range []string{"", "../escape", "0123456789abcdef0123456789abcdeg"} {
		if _, err := RPCSocketPath(invalid); err == nil {
			t.Fatal("invalid run ID accepted")
		}
	}
}
