package platform

import (
	"testing"
	"time"
)

func TestAwakeTimeAdvancesWhileRunning(t *testing.T) {
	start := time.Now()
	before, err := AwakeTime()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	after, err := AwakeTime()
	if err != nil {
		t.Fatal(err)
	}
	if delta := after - before; delta < 10*time.Millisecond || delta > time.Since(start)+100*time.Millisecond {
		t.Fatalf("invalid awake clock increment: %v", delta)
	}
}
