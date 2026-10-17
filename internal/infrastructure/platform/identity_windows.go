package platform

import (
	"errors"
	"strconv"
	"strings"

	"golang.org/x/sys/windows"
)

func ProcessIdentity(pid int) (Identity, error) {
	if pid <= 0 {
		return Identity{}, errors.New("platform: invalid PID")
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return Identity{}, err
	}
	defer windows.CloseHandle(h)
	return identityFromHandle(h, pid)
}

func identityFromHandle(h windows.Handle, pid int) (Identity, error) {
	var created, exited, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(h, &created, &exited, &kernel, &user); err != nil {
		return Identity{}, err
	}
	buffer := make([]uint16, 32768)
	size := uint32(len(buffer))
	if err := windows.QueryFullProcessImageName(h, 0, &buffer[0], &size); err != nil {
		return Identity{}, err
	}
	return Identity{PID: pid, Start: strconv.FormatUint(uint64(created.HighDateTime)<<32|uint64(created.LowDateTime), 10), Path: windows.UTF16ToString(buffer[:size])}, nil
}
func (i Identity) Alive() (bool, error) {
	h, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(i.PID))
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer windows.CloseHandle(h)
	result, err := windows.WaitForSingleObject(h, 0)
	if err != nil {
		return false, err
	}
	if result == windows.WAIT_OBJECT_0 {
		return false, nil
	}
	current, err := identityFromHandle(h, i.PID)
	if err != nil {
		// The process may exit between the wait and image-name query. Keep
		// using the same handle so PID reuse cannot affect this decision.
		if result, waitErr := windows.WaitForSingleObject(h, 0); waitErr == nil && result == windows.WAIT_OBJECT_0 {
			return false, nil
		}
		return false, err
	}
	return current.Start == i.Start && strings.EqualFold(current.Path, i.Path), nil
}
