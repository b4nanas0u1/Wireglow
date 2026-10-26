package platform

import (
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var queryUnbiasedInterruptTime = windows.NewLazySystemDLL("kernel32.dll").NewProc("QueryUnbiasedInterruptTime")

// AwakeTime is monotonic time since boot, excluding sleep and hibernation.
// Go's Windows time.Now uses interrupt time, which includes those intervals.
func AwakeTime() (time.Duration, error) {
	var ticks uint64
	ok, _, err := queryUnbiasedInterruptTime.Call(uintptr(unsafe.Pointer(&ticks)))
	if ok == 0 {
		return 0, err
	}
	return time.Duration(ticks) * 100 * time.Nanosecond, nil
}
