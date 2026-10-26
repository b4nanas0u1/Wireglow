package platform

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

func RequireManagedController() error {
	var inJob int32
	check := windows.NewLazySystemDLL("kernel32.dll").NewProc("IsProcessInJob")
	ok, _, err := check.Call(uintptr(windows.CurrentProcess()), 0, uintptr(unsafe.Pointer(&inJob)))
	if ok == 0 {
		return err
	}
	if inJob == 0 {
		return errors.New("platform: controller must be launched in Job Object")
	}
	return nil
}
