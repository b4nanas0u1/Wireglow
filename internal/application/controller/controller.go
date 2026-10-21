// Package controller управляет единственной сессией независимо от frontend и ОС.
package controller

import (
	"context"
	"errors"
	"sync"
	"time"

	"wireglow/internal/application/ports"
	"wireglow/internal/domain/profile"
	"wireglow/internal/domain/proxymode"
)

type State string

const (
	Idle              State = "Idle"
	Validating        State = "Validating"
	StartingRelay     State = "StartingRelay"
	StartingMihomo    State = "StartingMihomo"
	Probing           State = "Probing"
	ActivatingNetwork State = "ActivatingNetwork"
	Connected         State = "Connected"
	Stopping          State = "Stopping"
	Failed            State = "Failed"
)

var ErrClosed = errors.New("controller: closed")
var ErrBusy = errors.New("controller: session already active")

// Latency describes an HTTPS check through the active proxy. Times are Unix milliseconds.
type Latency struct {
	Checking  bool   `json:"checking"`
	DelayMs   int64  `json:"delayMs"`
	CheckedAt int64  `json:"checkedAt"`
	SuccessAt int64  `json:"successAt"`
	Error     string `json:"error,omitempty"`
}

type Snapshot struct {
	ActiveMode   proxymode.Mode `json:"activeMode,omitempty"`
	Latency      Latency        `json:"latency"`
	State        State          `json:"state"`
	Generation   uint64         `json:"generation"`
	ProfileID    string         `json:"profileId,omitempty"`
	MixedAddress string         `json:"mixedAddress,omitempty"`
	Error        string         `json:"error,omitempty"`
}

type Options struct {
	DefaultMode          proxymode.Mode
	NetworkForMode       func(proxymode.Mode) (ports.Network, error)
	StartupTimeout       time.Duration
	CleanupTimeout       time.Duration
	ManualCaptchaTimeout time.Duration
	LatencyInterval      time.Duration
}

type Controller struct {
	op             sync.Mutex
	mu             sync.Mutex
	runner         ports.Runner
	preparer       ports.Preparer
	network        ports.Network
	options        Options
	snapshot       Snapshot
	cancel         context.CancelFunc
	plan           *ports.Plan
	relay, engine  ports.Process
	closed         bool
	pendingStops   int
	latencyCtx     context.Context
	latencyCancel  context.CancelFunc
	measureLatency func(context.Context) (time.Duration, error)
	subs           map[chan Snapshot]struct{}
}

func New(r ports.Runner, p ports.Preparer, n ports.Network, o Options) *Controller {
	if o.LatencyInterval <= 0 {
		o.LatencyInterval = 30 * time.Second
	}
	if o.StartupTimeout <= 0 {
		o.StartupTimeout = 90 * time.Second
	}
	if o.CleanupTimeout <= 0 {
		o.CleanupTimeout = 20 * time.Second
	}
	if o.ManualCaptchaTimeout <= 0 {
		o.ManualCaptchaTimeout = 5 * time.Minute
	}
	return &Controller{runner: r, preparer: p, network: n, options: o, snapshot: Snapshot{State: Idle}, subs: make(map[chan Snapshot]struct{})}
}

func (c *Controller) Snapshot() Snapshot { c.mu.Lock(); defer c.mu.Unlock(); return c.snapshot }

// Subscribe выдаёт последний snapshot; медленный наблюдатель не задерживает cleanup.
func (c *Controller) Subscribe() (<-chan Snapshot, func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ch := make(chan Snapshot, 1)
	ch <- c.snapshot
	c.subs[ch] = struct{}{}
	var once sync.Once
	return ch, func() { once.Do(func() { c.mu.Lock(); delete(c.subs, ch); close(ch); c.mu.Unlock() }) }
}

func (c *Controller) set(state State, message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshot.State, c.snapshot.Error = state, message
	c.publishLocked()
}

func (c *Controller) publishLocked() {
	for ch := range c.subs {
		select {
		case <-ch:
		default:
		}
		ch <- c.snapshot
	}
}

// Connect ждёт прикладной проверки; Disconnect может отменить любую стадию запуска.
func (c *Controller) Connect(ctx context.Context, p profile.Profile) error {
	mode := c.options.DefaultMode
	if mode == "" {
		mode = proxymode.Local
	}
	return c.ConnectWithMode(ctx, p, mode)
}

func (c *Controller) ConnectWithMode(ctx context.Context, p profile.Profile, mode proxymode.Mode) error {
	if !mode.Valid() {
		return errors.New("controller: invalid proxy mode")
	}
	c.op.Lock()
	defer c.op.Unlock()
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClosed
	}
	if c.pendingStops > 0 {
		c.mu.Unlock()
		return ErrBusy
	}
	if c.plan != nil || c.snapshot.State == Connected {
		c.mu.Unlock()
		return ErrBusy
	}
	// Сетевой адаптер выбирается между сессиями под блокировкой операции.
	if c.options.NetworkForMode != nil {
		n, err := c.options.NetworkForMode(mode)
		if err != nil {
			c.mu.Unlock()
			return err
		}
		c.network = n
	}
	c.snapshot.ActiveMode = mode

	timeout := c.options.StartupTimeout
	if p.Connection.VK != nil && p.Connection.VK.ManualCaptcha && c.options.ManualCaptchaTimeout > timeout {
		timeout = c.options.ManualCaptchaTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	c.cancel = cancel
	c.snapshot.Generation++
	generation := c.snapshot.Generation
	c.snapshot.ProfileID = p.ID
	c.snapshot.Latency = Latency{}
	c.mu.Unlock()
	defer func() { cancel(); c.mu.Lock(); c.cancel = nil; c.mu.Unlock() }()
	c.set(Validating, "")
	if err := c.network.Recover(ctx); err != nil {
		return c.fail("network recovery failed", err)
	}
	var plan *ports.Plan
	var err error
	if preparer, ok := c.preparer.(ports.ModePreparer); ok {
		plan, err = preparer.PrepareForMode(ctx, p, mode)
	} else {
		plan, err = c.preparer.Prepare(ctx, p)
	}
	if err != nil {
		return c.fail("connection preparation failed", err)
	}
	if mode == proxymode.Tun && !plan.TunReady {
		_ = plan.Close()
		return c.fail("TUN bootstrap unavailable", errors.New("tun: reconnect with TUN bootstrap"))
	}
	c.plan = plan
	c.set(StartingRelay, "")
	if c.relay, err = c.runner.Start(ctx, plan.Relay); err != nil {
		return c.fail("relay start failed", err)
	}
	// Выход любого ядра отменяет ожидание readiness/probe.
	go cancelOnExit(ctx, c.relay, cancel)
	c.set(StartingMihomo, "")
	if c.engine, err = c.runner.Start(ctx, plan.Engine); err != nil {
		return c.fail("mihomo start failed", err)
	}
	go cancelOnExit(ctx, c.engine, cancel)
	if err = plan.Ready(ctx); err != nil {
		return c.fail("mihomo readiness failed", err)
	}
	c.set(Probing, "")
	if err = waitOutbound(ctx, plan.Probe); err != nil {
		return c.fail("outbound probe failed", err)
	}
	if diagnostic, ok := c.engine.(ports.ProcessDiagnostic); ok {
		diagnostic.ResetProcessDiagnostic()
	}
	c.set(ActivatingNetwork, "")
	if err = c.network.Apply(ctx, plan.MixedAddress); err != nil {
		return c.fail("network activation failed", err)
	}
	if err = plan.Probe(ctx); err != nil {
		return c.fail("probe after activation failed", err)
	}
	if err = ctx.Err(); err != nil {
		return c.fail("connection cancelled", err)
	}
	for _, process := range []ports.Process{c.relay, c.engine} {
		select {
		case <-process.Done():
			return c.fail("engine exited during startup", process.Err())
		default:
		}
	}
	c.mu.Lock()
	c.snapshot.MixedAddress = plan.MixedAddress
	c.mu.Unlock()
	if plan.MeasureLatency != nil {
		c.mu.Lock()
		latencyCtx, latencyCancel := context.WithCancel(context.Background())
		c.latencyCtx, c.latencyCancel, c.measureLatency = latencyCtx, latencyCancel, plan.MeasureLatency
		c.mu.Unlock()
		c.set(Connected, "")
		go c.pollLatency(latencyCtx)
	} else {
		c.set(Connected, "")
	}
	go c.monitor(generation, c.relay, c.engine)
	c.watchNetwork(generation, c.network, c.relay)
	return nil
}

// API mihomo готов раньше relay: VK ещё может получать TURN-реквизиты и решать CAPTCHA.
func waitOutbound(ctx context.Context, probe func(context.Context) error) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := probe(ctx)
		if err == nil {
			return ctx.Err()
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return errors.Join(err, ctx.Err())
		case <-timer.C:
		}
	}
}

func cancelOnExit(ctx context.Context, p ports.Process, cancel context.CancelFunc) {
	select {
	case <-ctx.Done():
	case <-p.Done():
		cancel()
	}
}

func (c *Controller) monitor(generation uint64, relay, engine ports.Process) {
	select {
	case <-relay.Done():
	case <-engine.Done():
	}
	c.op.Lock()
	defer c.op.Unlock()
	if c.Snapshot().Generation != generation || c.plan == nil {
		return
	}
	_ = c.fail("engine exited", errors.New("engine exited"))
}

func (c *Controller) fail(message string, cause error) error {
	networkFailure := message == "network activation failed" || message == "probe after activation failed"
	engine := c.engine
	var diagnostic ports.DiagnosticError
	if errors.As(cause, &diagnostic) {
		message += ": " + diagnostic.PublicDiagnostic()
	}
	c.set(Stopping, "")
	err := c.cleanup()
	if diagnostic, ok := c.network.(ports.NetworkDiagnostic); ok && networkFailure {
		if detail := diagnostic.NetworkDiagnostic(); detail != "" {
			message += ": " + detail
		}
	}
	if err != nil {
		message += "; cleanup incomplete"
	}
	if process, ok := engine.(ports.ProcessDiagnostic); ok {
		if detail := process.ProcessDiagnostic(); detail != "" {
			message += "; mihomo: " + detail
		} else if networkFailure {
			message += "; mihomo: no outbound warnings"
		}
	}
	c.set(Failed, message)
	return errors.Join(cause, err)
}

func (c *Controller) cleanup() error {
	c.mu.Lock()
	if c.latencyCancel != nil {
		c.latencyCancel()
	}
	c.latencyCtx, c.latencyCancel, c.measureLatency = nil, nil, nil
	c.snapshot.Latency = Latency{}
	c.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), c.options.CleanupTimeout)
	defer cancel()
	// Ошибка восстановления не препятствует остановке процессов; журнал остаётся для повтора.
	netErr := c.network.Recover(ctx)
	var stopErr error
	for _, p := range []ports.Process{c.engine, c.relay} {
		if p != nil {
			stopErr = errors.Join(stopErr, p.Stop(ctx))
		}
	}
	if stopErr != nil {
		return errors.Join(netErr, stopErr)
	}
	c.engine, c.relay = nil, nil
	if c.plan != nil {
		if err := c.plan.Close(); err != nil {
			return errors.Join(netErr, err)
		}
		c.plan = nil
	}
	c.mu.Lock()
	c.snapshot.MixedAddress = ""
	c.mu.Unlock()
	return netErr
}

func (c *Controller) Disconnect() error {
	c.mu.Lock()
	c.pendingStops++
	if c.cancel != nil {
		c.cancel()
	}
	c.mu.Unlock()
	c.op.Lock()
	defer func() { c.mu.Lock(); c.pendingStops--; c.mu.Unlock(); c.op.Unlock() }()
	c.set(Stopping, "")
	if err := c.cleanup(); err != nil {
		c.set(Failed, "cleanup incomplete")
		return err
	}
	c.mu.Lock()
	c.snapshot.ProfileID = ""
	c.snapshot.ActiveMode = ""
	c.mu.Unlock()
	c.set(Idle, "")
	return nil
}

func (c *Controller) Shutdown() error {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	return c.Disconnect()
}

// WatchOwner — отдельный lease владельца; subscriptions не продлевают сессию.
// now must be monotonic and exclude system suspend. Timer delivery alone is
// not proof of owner loss: Windows Go timers include time spent asleep.
func (c *Controller) WatchOwner(ctx context.Context, beats <-chan struct{}, timeout time.Duration, now func() (time.Duration, error)) error {
	if timeout <= 0 || now == nil {
		return errors.New("controller: invalid owner timeout")
	}
	lastBeat, err := now()
	if err != nil {
		return errors.Join(err, c.Shutdown())
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return c.Shutdown()
		case _, ok := <-beats:
			if !ok {
				return c.Shutdown()
			}
			lastBeat, err = now()
			if err != nil {
				return errors.Join(err, c.Shutdown())
			}
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(timeout)
		case <-timer.C:
			current, err := now()
			if err != nil {
				return errors.Join(err, c.Shutdown())
			}
			remaining := timeout - (current - lastBeat)
			if remaining <= 0 {
				return c.Shutdown()
			}
			// Resume may deliver an expired Go timer before the launcher has
			// published its next heartbeat. Preserve the remaining awake lease.
			timer.Reset(remaining)
		}
	}
}

// RelayPID is used only to authenticate the relay's process identity to TUN.
// Session operations already serialize process replacement through op.
func (c *Controller) RelayPID() int {
	if p, ok := c.relay.(interface{ PID() int }); ok {
		return p.PID()
	}
	return 0
}
func (c *Controller) watchNetwork(generation uint64, n ports.Network, relay ports.Process) {
	health, ok := n.(interface{ Failed() <-chan struct{} })
	if !ok {
		return
	}
	done := health.Failed()
	if done == nil {
		return
	}
	go func() {
		select {
		case <-done:
		case <-relay.Done():
			return
		}
		c.op.Lock()
		defer c.op.Unlock()
		if c.Snapshot().Generation != generation || c.network != n || c.plan == nil {
			return
		}
		_ = c.fail("TUN helper lost; connection stopped", errors.New("tun: helper lost"))
	}()
}
