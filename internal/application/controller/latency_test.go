package controller

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"wireglow/internal/domain/profile"
)

func awaitLatency(t *testing.T, c *Controller, predicate func(Latency) bool) Latency {
	t.Helper()
	timeout := time.After(time.Second)
	updates, unsubscribe := c.Subscribe()
	defer unsubscribe()
	for {
		select {
		case s := <-updates:
			if predicate(s.Latency) {
				return s.Latency
			}
		case <-timeout:
			t.Fatalf("latency not published: %+v", c.Snapshot())
			return Latency{}
		}
	}
}

func TestLatencyFailurePreservesLastSuccessAndSession(t *testing.T) {
	c, _, _, plan, _ := harness()
	var failing atomic.Bool
	plan.MeasureLatency = func(context.Context) (time.Duration, error) {
		if failing.Load() {
			return 0, errors.New("secret endpoint")
		}
		return 123 * time.Millisecond, nil
	}
	if err := c.Connect(context.Background(), profile.Profile{ID: "one"}); err != nil {
		t.Fatal(err)
	}
	defer c.Disconnect()
	first := awaitLatency(t, c, func(l Latency) bool { return l.SuccessAt > 0 && !l.Checking })
	if first.DelayMs != 123 {
		t.Fatal(first)
	}
	failing.Store(true)
	if err := c.CheckLatency(context.Background()); err != nil {
		t.Fatal(err)
	}
	last := c.Snapshot()
	if last.State != Connected || last.Latency.Error != "HTTPS check failed" || last.Latency.SuccessAt != first.SuccessAt || last.Latency.DelayMs != 123 {
		t.Fatal(last)
	}
}

func TestLatencyCancelledOnDisconnectAndNoStaleResult(t *testing.T) {
	c, _, _, plan, _ := harness()
	entered := make(chan struct{})
	finished := make(chan struct{})
	plan.MeasureLatency = func(ctx context.Context) (time.Duration, error) {
		close(entered)
		<-ctx.Done()
		close(finished)
		return time.Second, nil
	}
	if err := c.Connect(context.Background(), profile.Profile{ID: "one"}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("check not started")
	}
	if err := c.CheckLatency(context.Background()); !errors.Is(err, ErrBusy) {
		t.Fatal(err)
	}
	if err := c.Disconnect(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("check not cancelled")
	}
	if s := c.Snapshot(); s.State != Idle || s.Latency != (Latency{}) {
		t.Fatal(s)
	}
	if err := c.CheckLatency(context.Background()); !errors.Is(err, ErrBusy) {
		t.Fatal(err)
	}
}

func TestPeriodicLatency(t *testing.T) {
	c, _, _, plan, _ := harness()
	c.options.LatencyInterval = 10 * time.Millisecond
	var calls atomic.Int32
	plan.MeasureLatency = func(context.Context) (time.Duration, error) { calls.Add(1); return time.Millisecond, nil }
	if err := c.Connect(context.Background(), profile.Profile{ID: "one"}); err != nil {
		t.Fatal(err)
	}
	defer c.Disconnect()
	awaitLatency(t, c, func(l Latency) bool { return calls.Load() >= 2 && !l.Checking })
}
