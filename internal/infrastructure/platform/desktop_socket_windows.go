package platform

import "path/filepath"

// Keep the existing Windows endpoint and DACL behavior unchanged.
func DesktopSocket(runDir, _ string) (string, error) {
	return filepath.Join(runDir, "control.sock"), nil
}
func CleanupDesktopSocket(string) error { return nil }
