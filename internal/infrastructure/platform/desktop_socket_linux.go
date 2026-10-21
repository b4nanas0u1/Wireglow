package platform

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// Only transient IPC leaves the portable directory. Request, lease, identity,
// locks and recovery journals remain beside the portable user data.
func desktopSocketPath(runID string) (string, error) {
	id, err := hex.DecodeString(runID)
	if err != nil || len(id) != 16 {
		return "", errors.New("platform: invalid IPC run ID")
	}
	for _, base := range []string{os.Getenv("XDG_RUNTIME_DIR"), filepath.Join("/run/user", strconv.Itoa(os.Getuid()))} {
		if !filepath.IsAbs(base) {
			continue
		}
		socket := filepath.Join(base, "wg-"+runID, "control.sock")
		if len(socket) >= 108 {
			continue
		}
		info, e := os.Lstat(base)
		if e != nil || !info.IsDir() || info.Mode().Perm()&0077 != 0 {
			continue
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok || int(stat.Uid) != os.Getuid() {
			continue
		}
		return socket, nil
	}
	return "", errors.New("platform: short private system runtime required for controller IPC")
}
func DesktopSocket(_ string, runID string) (string, error) {
	socket, err := desktopSocketPath(runID)
	if err != nil {
		return "", err
	}
	// Never reuse a directory or socket left by another run.
	if err = os.Mkdir(filepath.Dir(socket), 0700); err != nil {
		return "", err
	}
	return socket, nil
}
func CleanupDesktopSocket(runID string) error {
	socket, err := desktopSocketPath(runID)
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Dir(socket))
}
