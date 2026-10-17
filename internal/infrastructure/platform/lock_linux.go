package platform

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

var ErrLocked = errors.New("platform: another controller is running")

type Lock struct{ file *os.File }

func AcquireLock(path string) (*Lock, error) {
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), path)
	if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, ErrLocked
		}
		return nil, err
	}
	return &Lock{f}, nil
}
func (l *Lock) Close() error { return l.file.Close() }
