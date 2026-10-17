package controller

import (
	"context"
	"errors"

	"wireglow/internal/domain/proxymode"
)

var ErrModeSwitch = errors.New("controller: proxy mode switch failed")

// SwitchMode меняет только настройки ОС, сохраняя процессы, порты и поколение сессии.
func (c *Controller) SwitchMode(ctx context.Context, mode proxymode.Mode) error {
	if !mode.Valid() {
		return ErrModeSwitch
	}
	c.op.Lock()
	defer c.op.Unlock()
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClosed
	}
	if c.pendingStops > 0 || c.snapshot.State != Connected || c.plan == nil {
		c.mu.Unlock()
		return ErrBusy
	}
	if c.snapshot.ActiveMode == mode {
		c.mu.Unlock()
		return nil
	}
	address := c.snapshot.MixedAddress
	c.mu.Unlock()
	if c.options.NetworkForMode == nil {
		return ErrModeSwitch
	}
	target, err := c.options.NetworkForMode(mode)
	if err != nil {
		return errors.Join(ErrModeSwitch, err)
	}
	ctx, cancel := context.WithTimeout(ctx, c.options.CleanupTimeout)
	defer cancel()
	// Восстанавливаем прежние настройки до включения другого режима.
	if err := c.network.Recover(ctx); err != nil {
		c.modeResult("", "proxy mode switch failed; network recovery incomplete")
		return errors.Join(ErrModeSwitch, err)
	}
	// Адаптер остаётся ответственным за журнал даже после частичного Apply.
	c.network = target
	if err := target.Apply(ctx, address); err != nil {
		// Отмена RPC не должна мешать откату уже записанных изменений ОС.
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), c.options.CleanupTimeout)
		defer cleanupCancel()
		rollbackErr := target.Recover(cleanupCtx)
		if rollbackErr != nil {
			c.modeResult("", "proxy mode switch failed; network recovery incomplete")
		} else {
			c.modeResult(proxymode.Local, "proxy mode switch failed; system settings restored")
		}
		return errors.Join(ErrModeSwitch, err, rollbackErr)
	}
	c.modeResult(mode, "")
	return nil
}

func (c *Controller) modeResult(mode proxymode.Mode, message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshot.ActiveMode = mode
	c.snapshot.Error = message
	c.publishLocked()
}
