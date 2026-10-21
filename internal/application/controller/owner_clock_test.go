package controller

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"wireglow/internal/domain/profile"
)

func TestOwnerLeaseExcludesSleepButStillExpires(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c, _, _, _, _ := harness()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		beats := make(chan struct{})
		var awake atomic.Int64
		done := make(chan error, 1)
		go func() {
			done <- c.WatchOwner(ctx, beats, 8*time.Second, func() (time.Duration, error) { return time.Duration(awake.Load()), nil })
		}()
		synctest.Wait()
		// Simulate a Windows suspend: Go time advances but awake time does not.
		time.Sleep(time.Hour)
		synctest.Wait()
		select {
		case err := <-done:
			t.Fatalf("sleep closed a live owner's controller: %v", err)
		default:
		}
		awake.Store(int64(7 * time.Second))
		beats <- struct{}{}
		synctest.Wait()
		awake.Store(int64(14 * time.Second))
		time.Sleep(8 * time.Second)
		synctest.Wait()
		select {
		case err := <-done:
			t.Fatalf("heartbeat did not renew awake deadline: %v", err)
		default:
		}
		awake.Store(int64(15 * time.Second))
		time.Sleep(8 * time.Second)
		synctest.Wait()
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatal("hung owner did not expire after 8 seconds awake")
		}
		if err := c.Connect(ctx, profile.Profile{}); !errors.Is(err, ErrClosed) {
			t.Fatalf("expired controller accepts commands: %v", err)
		}
	})
}

func TestOwnerClockFailureClosesController(t *testing.T) {
	c, _, _, _, _ := harness()
	clockErr := errors.New("clock unavailable")
	err := c.WatchOwner(context.Background(), make(chan struct{}), time.Second, func() (time.Duration, error) {
		return 0, clockErr
	})
	if !errors.Is(err, clockErr) {
		t.Fatalf("clock failure hidden: %v", err)
	}
	if err := c.Connect(context.Background(), profile.Profile{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("clock failure left controller running: %v", err)
	}
}
