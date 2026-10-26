package controller

import (
	"context"
	"time"
)

func (c *Controller) pollLatency(ctx context.Context) {
	_ = c.CheckLatency(ctx)
	ticker := time.NewTicker(c.options.LatencyInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = c.CheckLatency(ctx)
		}
	}
}

// CheckLatency serializes checks without holding the lifecycle lock during I/O.
// Network failures are published as telemetry, not as session failures.
func (c *Controller) CheckLatency(ctx context.Context) error {
	c.mu.Lock()
	if err := ctx.Err(); err != nil {
		c.mu.Unlock()
		return err
	}
	if c.snapshot.State != Connected || c.measureLatency == nil || c.snapshot.Latency.Checking || c.latencyCtx.Err() != nil {
		c.mu.Unlock()
		return ErrBusy
	}
	session := c.latencyCtx
	generation := c.snapshot.Generation
	measure := c.measureLatency
	c.snapshot.Latency.Checking = true
	c.publishLocked()
	c.mu.Unlock()

	checkCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	stop := context.AfterFunc(session, cancel)
	defer stop()
	defer cancel()
	duration, err := measure(checkCtx)
	if err == nil {
		err = checkCtx.Err()
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if session.Err() != nil || c.snapshot.Generation != generation || c.snapshot.State != Connected {
		return nil
	}
	c.snapshot.Latency.Checking = false
	c.snapshot.Latency.CheckedAt = time.Now().UnixMilli()
	if err != nil {
		c.snapshot.Latency.Error = "HTTPS check failed"
	} else {
		c.snapshot.Latency.DelayMs = max(1, duration.Milliseconds())
		c.snapshot.Latency.SuccessAt = c.snapshot.Latency.CheckedAt
		c.snapshot.Latency.Error = ""
	}
	c.publishLocked()
	return nil
}
