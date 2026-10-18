// Package platform содержит адаптеры ОС, используемые Go launcher и controller.
package platform

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Paths struct {
	Data    string
	Runtime string
}

func DefaultPaths() (Paths, error) {
	executable, err := os.Executable()
	if err != nil {
		return Paths{}, err
	}
	if paths, portable, err := portablePaths(executable); err != nil || portable {
		return paths, err
	}
	config, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, err
	}
	return makePaths(config)
}

// Controller lives in bin/ in a distribution. Also support a standalone binary.
func portablePaths(executable string) (Paths, bool, error) {
	root := filepath.Dir(executable)
	if filepath.Base(root) == "bin" {
		root = filepath.Dir(root)
	}
	marker, err := os.Stat(filepath.Join(root, "portable.json"))
	if os.IsNotExist(err) {
		return Paths{}, false, nil
	}
	if err != nil {
		return Paths{}, false, err
	}
	if !marker.Mode().IsRegular() {
		return Paths{}, false, errors.New("platform: portable.json must be a regular file")
	}
	data := filepath.Join(root, "data")
	return Paths{Data: data, Runtime: filepath.Join(data, "runtime")}, true, nil
}

func ensureDir(path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("platform: unsafe directory")
	}
	return protectDir(path)
}

func EnsurePaths(p Paths) error {
	for _, path := range []string{p.Data, p.Runtime, filepath.Join(p.Data, "profiles")} {
		if !filepath.IsAbs(path) {
			return errors.New("platform: absolute directories required")
		}
		if err := ensureDir(path); err != nil {
			return err
		}
	}
	return nil
}

func ProxyPort(address string) (int, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil || host != "127.0.0.1" {
		return 0, errors.New("platform: proxy must use IPv4 loopback")
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return 0, errors.New("platform: invalid port")
	}
	return n, nil
}

func pause(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
