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
	config, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, err
	}
	return makePaths(config)
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
