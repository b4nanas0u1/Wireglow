package platform

import "os"

func ipcRuntimeRoot() string { return os.TempDir() }
