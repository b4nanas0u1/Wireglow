package platform

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDesktopSocketIndependentOfLongPortablePath(t *testing.T) {
	base, err := os.MkdirTemp("/tmp", "wg-ipc-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(base)
	t.Setenv("XDG_RUNTIME_DIR", base)
	id := "0123456789abcdef0123456789abcdef"
	socket, err := DesktopSocket(filepath.Join(t.TempDir(), strings.Repeat("long-portable-path", 20)), id)
	if err != nil {
		t.Fatal(err)
	}
	if len(socket) >= 108 || filepath.Dir(filepath.Dir(socket)) != base {
		t.Fatal("unsafe socket path")
	}
	defer CleanupDesktopSocket(id)
	info, err := os.Stat(filepath.Dir(socket))
	if err != nil || info.Mode().Perm() != 0700 {
		t.Fatal("IPC directory not private", err)
	}
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	listener.Close()
	if _, err = DesktopSocket("", id); err == nil {
		t.Fatal("reused another run directory")
	}
	if err = CleanupDesktopSocket(id); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Dir(socket)); !os.IsNotExist(err) {
		t.Fatal("IPC directory retained")
	}
	if _, err = DesktopSocket("", "../other"); err == nil {
		t.Fatal("invalid run ID accepted")
	}
}
