package platform

import (
	"time"

	"golang.org/x/sys/unix"
)

// AwakeTime uses CLOCK_MONOTONIC, which excludes system suspend on Linux.
func AwakeTime() (time.Duration, error) {
	var ts unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_MONOTONIC, &ts); err != nil {
		return 0, err
	}
	return time.Duration(ts.Nano()), nil
}
