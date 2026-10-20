package platform

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
)

// RPCSocketPath is independent of the bundle and persistent recovery state.
// Both launcher children inherit the same per-user runtime environment.
func RPCSocketPath(runID string) (string, error) {
	if decoded, err := hex.DecodeString(runID); err != nil || len(decoded) != 16 {
		return "", errors.New("platform: invalid RPC run ID")
	}
	base := ipcRuntimeRoot()
	if !filepath.IsAbs(base) {
		return "", errors.New("platform: absolute IPC runtime directory required")
	}
	return filepath.Join(base, "wg-"+runID, "rpc"), nil
}

// PrepareRPCSocket runs in the independent watchdog before it announces readiness.
// Its cleanup therefore also survives a crash of the GUI and controller Job.
func PrepareRPCSocket(runID string) (func() error, error) {
	path, err := RPCSocketPath(runID)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(path)
	if err := os.Mkdir(dir, 0700); err != nil {
		return nil, err
	}
	remove := func(path string) error {
		err := os.Remove(path)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	cleanup := func() error { return errors.Join(remove(path), remove(dir)) }
	if err := protectDir(dir); err != nil {
		return nil, errors.Join(err, cleanup())
	}
	return cleanup, nil
}
