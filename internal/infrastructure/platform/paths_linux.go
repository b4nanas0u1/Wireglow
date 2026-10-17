package platform

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/sys/unix"
)

func makePaths(config string) (Paths, error) {
	runtime := os.Getenv("XDG_RUNTIME_DIR")
	if runtime == "" {
		runtime = filepath.Join("/run/user", strconv.Itoa(os.Getuid()))
	}
	return Paths{filepath.Join(config, "wireglow"), filepath.Join(runtime, "wireglow")}, nil
}

func protectDir(path string) error {
	var stat unix.Stat_t
	if err := unix.Lstat(path, &stat); err != nil {
		return err
	}
	if int(stat.Uid) != os.Getuid() {
		return errors.New("platform: directory belongs to another user")
	}
	return os.Chmod(path, 0700)
}
