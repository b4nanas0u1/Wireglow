package platform

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

var ErrLocked = errors.New("platform: another controller is running")

type Lock struct {
	file       *os.File
	overlapped windows.Overlapped
}

func AcquireLock(path string) (*Lock, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	l := &Lock{file: f}
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &l.overlapped); err != nil {
		f.Close()
		if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return nil, ErrLocked
		}
		return nil, err
	}
	return l, nil
}
func (l *Lock) Close() error { return l.file.Close() }
