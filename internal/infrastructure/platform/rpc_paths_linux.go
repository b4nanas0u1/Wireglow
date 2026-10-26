package platform

import (
	"os"
	"path/filepath"
	"strconv"
)

func ipcRuntimeRoot() string {
	if base := os.Getenv("XDG_RUNTIME_DIR"); base != "" {
		return base
	}
	return filepath.Join("/run/user", strconv.Itoa(os.Getuid()))
}
