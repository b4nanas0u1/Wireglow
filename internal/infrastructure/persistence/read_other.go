//go:build !windows

package persistence

import "os"

func openReadFile(path string) (*os.File, error) { return os.Open(path) }
