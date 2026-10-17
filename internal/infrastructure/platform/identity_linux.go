package platform

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

func ProcessIdentity(pid int) (Identity, error) {
	if pid <= 0 {
		return Identity{}, errors.New("platform: invalid PID")
	}
	base := "/proc/" + strconv.Itoa(pid)
	data, err := os.ReadFile(base + "/stat")
	if err != nil {
		return Identity{}, err
	}
	fields := strings.Fields(string(data)[strings.LastIndex(string(data), ")")+1:])
	if len(fields) < 20 {
		return Identity{}, errors.New("platform: invalid process stat")
	}
	path, err := os.Readlink(base + "/exe")
	if err != nil {
		return Identity{}, err
	}
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return Identity{}, err
	}
	return Identity{pid, fields[19], path, strings.TrimSpace(string(boot))}, nil
}

func (i Identity) Alive() (bool, error) {
	current, err := ProcessIdentity(i.PID)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return current == i, nil
}
