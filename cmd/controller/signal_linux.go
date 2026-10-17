package main

import (
	"context"
	"os/signal"
	"syscall"
)

func terminationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, syscall.SIGTERM)
}
