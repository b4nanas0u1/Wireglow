package controller

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"wireglow/internal/application/ports"
	"wireglow/internal/domain/profile"
)

type trace struct {
	mu    sync.Mutex
	items []string
}

func (r *trace) add(s string) { r.mu.Lock(); r.items = append(r.items, s); r.mu.Unlock() }
func (r *trace) text() string { r.mu.Lock(); defer r.mu.Unlock(); return strings.Join(r.items, ",") }

type fakeProcess struct {
	done    chan struct{}
	once    sync.Once
	name    string
	trace   *trace
	stopErr error
}

func (p *fakeProcess) Done() <-chan struct{} { return p.done }
func (p *fakeProcess) Err() error            { return errors.New("child exited") }
func (p *fakeProcess) exit()                 { p.once.Do(func() { close(p.done) }) }
func (p *fakeProcess) Stop(ctx context.Context) error {
	p.trace.add("stop:" + p.name)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if p.stopErr != nil {
		return p.stopErr
	}
	p.exit()
	return nil
}

type fakeRunner struct {
	trace     *trace
	processes []*fakeProcess
	fail      string
}

func (r *fakeRunner) Start(ctx context.Context, c ports.Command) (ports.Process, error) {
	r.trace.add("start:" + c.Path)
	if c.Path == r.fail {
		return nil, errors.New("start failed")
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	p := &fakeProcess{done: make(chan struct{}), name: c.Path, trace: r.trace}
	r.processes = append(r.processes, p)
	return p, nil
}

type fakePreparer struct {
	plan *ports.Plan
	err  error
}

func (p fakePreparer) Prepare(ctx context.Context, _ profile.Profile) (*ports.Plan, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return p.plan, p.err
}

type fakeNetwork struct {
	trace                *trace
	recoverErr, applyErr error
}

func (n *fakeNetwork) Recover(ctx context.Context) error {
	n.trace.add("recover")
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return n.recoverErr
}
func (n *fakeNetwork) Apply(ctx context.Context, _ string) error {
	n.trace.add("apply")
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return n.applyErr
}
func harness() (*Controller, *fakeRunner, *fakeNetwork, *ports.Plan, *trace) {
	tr := &trace{}
	runner := &fakeRunner{trace: tr}
	n := &fakeNetwork{trace: tr}
	plan := &ports.Plan{Relay: ports.Command{Path: "relay"}, Engine: ports.Command{Path: "engine"}, MixedAddress: "127.0.0.1:1234", Ready: func(context.Context) error { tr.add("ready"); return nil }, Probe: func(context.Context) error { tr.add("probe"); return nil }, Close: func() error { tr.add("close"); return nil }}
	c := New(runner, fakePreparer{plan: plan}, n, Options{StartupTimeout: time.Second, CleanupTimeout: time.Second})
	return c, runner, n, plan, tr
}

func TestConnectDisconnectOrder(t *testing.T) {
	c, runner, _, _, tr := harness()
	if err := c.Connect(context.Background(), profile.Profile{ID: "one"}); err != nil {
		t.Fatal(err)
	}
	if s := c.Snapshot(); s.State != Connected || s.ProfileID != "one" || s.MixedAddress == "" {
		t.Fatalf("snapshot: %+v", s)
	}
	if err := c.Connect(context.Background(), profile.Profile{ID: "two"}); !errors.Is(err, ErrBusy) {
		t.Fatal(err)
	}
	if len(runner.processes) != 2 {
		t.Fatal("duplicate processes")
	}
	if err := c.Disconnect(); err != nil {
		t.Fatal(err)
	}
	const want = "recover,start:relay,start:engine,ready,probe,apply,probe,recover,stop:engine,stop:relay,close"
	if tr.text() != want {
		t.Fatalf("order = %s", tr.text())
	}
	if err := c.Disconnect(); err != nil {
		t.Fatal(err)
	}
	if c.Snapshot().State != Idle {
		t.Fatal(c.Snapshot())
	}
}

func TestRollbackOnEveryStartupFailure(t *testing.T) {
	for _, stage := range []string{"prepare", "relay", "engine", "ready", "probe", "apply", "post-probe"} {
		t.Run(stage, func(t *testing.T) {
			c, runner, n, plan, tr := harness()
			failure := errors.New("sensitive-input")
			switch stage {
			case "prepare":
				c.preparer = fakePreparer{err: failure}
			case "relay":
				runner.fail = "relay"
			case "engine":
				runner.fail = "engine"
			case "ready":
				plan.Ready = func(context.Context) error { return failure }
			case "probe":
				plan.Probe = func(context.Context) error { return failure }
			case "apply":
				n.applyErr = failure
			case "post-probe":
				calls := 0
				plan.Probe = func(context.Context) error {
					calls++
					if calls == 2 {
						return failure
					}
					return nil
				}
			}
			if err := c.Connect(context.Background(), profile.Profile{}); err == nil {
				t.Fatal("expected failure")
			}
			if s := c.Snapshot(); s.State != Failed || strings.Contains(s.Error, "sensitive-input") {
				t.Fatal(s)
			}
			for _, p := range runner.processes {
				select {
				case <-p.Done():
				default:
					t.Fatal("surviving process")
				}
			}
			if strings.Contains(tr.text(), "stop:engine,stop:relay") && !strings.Contains(tr.text(), "recover,stop:engine") {
				t.Fatal("network must restore before stop")
			}
		})
	}
}

func TestDisconnectCancelsStartup(t *testing.T) {
	for _, stage := range []string{"ready", "probe", "apply"} {
		t.Run(stage, func(t *testing.T) {
			c, _, _, plan, _ := harness()
			entered := make(chan struct{})
			wait := func(ctx context.Context) error { close(entered); <-ctx.Done(); return ctx.Err() }
			switch stage {
			case "ready":
				plan.Ready = wait
			case "probe":
				plan.Probe = wait
			case "apply":
				c.network = &blockingNetwork{wait: wait}
			}
			done := make(chan error, 1)
			go func() { done <- c.Connect(context.Background(), profile.Profile{}) }()
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("stage not reached")
			}
			if err := c.Disconnect(); err != nil {
				t.Fatal(err)
			}
			if err := <-done; err == nil {
				t.Fatal("startup succeeded after cancellation")
			}
			if c.Snapshot().State != Idle {
				t.Fatal(c.Snapshot())
			}
		})
	}
}

func TestConnectWaitsForOutboundBeforeNetworkActivation(t *testing.T) {
	c, _, _, plan, tr := harness()
	c.options.StartupTimeout = 3 * time.Second
	calls := 0
	plan.Probe = func(context.Context) error {
		calls++
		tr.add("probe")
		if calls == 1 {
			return errors.New("relay still authorizing")
		}
		return nil
	}
	defer c.Disconnect()
	if err := c.Connect(context.Background(), profile.Profile{}); err != nil {
		t.Fatal(err)
	}
	if c.Snapshot().State != Connected || calls != 3 {
		t.Fatalf("snapshot=%+v probes=%d", c.Snapshot(), calls)
	}
	if got := tr.text(); got != "recover,start:relay,start:engine,ready,probe,probe,apply,probe" {
		t.Fatalf("order = %s", got)
	}
}

func TestOutboundRetryStopsOnCancellationOrEngineExit(t *testing.T) {
	for _, engineExit := range []bool{false, true} {
		t.Run(map[bool]string{false: "disconnect", true: "engine exit"}[engineExit], func(t *testing.T) {
			c, runner, _, plan, tr := harness()
			c.options.StartupTimeout = 10 * time.Second
			entered := make(chan struct{})
			calls := 0
			plan.Probe = func(context.Context) error {
				calls++
				if calls == 1 {
					if engineExit {
						runner.processes[0].exit()
					}
					close(entered)
				}
				return errors.New("relay not ready")
			}
			done := make(chan error, 1)
			go func() { done <- c.Connect(context.Background(), profile.Profile{}) }()
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("probe not reached")
			}
			if !engineExit {
				if err := c.Disconnect(); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("error = %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("retry did not stop promptly")
			}
			if calls != 1 || strings.Contains(tr.text(), "apply") {
				t.Fatalf("probes=%d trace=%s", calls, tr.text())
			}
			for _, p := range runner.processes {
				select {
				case <-p.Done():
				default:
					t.Fatal("surviving process")
				}
			}
		})
	}
}

func TestOutboundFailureRespectsStartupDeadline(t *testing.T) {
	c, runner, _, plan, tr := harness()
	c.options.StartupTimeout = 30 * time.Millisecond
	failure := errors.New("relay unavailable")
	plan.Probe = func(context.Context) error { return failure }
	err := c.Connect(context.Background(), profile.Profile{})
	if !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, failure) {
		t.Fatalf("error = %v", err)
	}
	if c.Snapshot().State != Failed || strings.Contains(tr.text(), "apply") {
		t.Fatalf("snapshot=%+v trace=%s", c.Snapshot(), tr.text())
	}
	for _, p := range runner.processes {
		select {
		case <-p.Done():
		default:
			t.Fatal("surviving process")
		}
	}
}

type blockingNetwork struct{ wait func(context.Context) error }

func (n *blockingNetwork) Recover(context.Context) error             { return nil }
func (n *blockingNetwork) Apply(ctx context.Context, _ string) error { return n.wait(ctx) }

func TestCrashRollsBackBothEngines(t *testing.T) {
	for _, index := range []int{0, 1} {
		t.Run(string(rune('0'+index)), func(t *testing.T) {
			c, r, _, _, _ := harness()
			if err := c.Connect(context.Background(), profile.Profile{}); err != nil {
				t.Fatal(err)
			}
			r.processes[index].exit()
			deadline := time.Now().Add(time.Second)
			for c.Snapshot().State != Failed && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
			}
			if c.Snapshot().State != Failed {
				t.Fatal(c.Snapshot())
			}
			for _, p := range r.processes {
				select {
				case <-p.Done():
				default:
					t.Fatal("pair not stopped")
				}
			}
		})
	}
}

func TestEngineExitCancelsReadiness(t *testing.T) {
	c, r, _, plan, _ := harness()
	plan.Ready = func(ctx context.Context) error { r.processes[0].exit(); <-ctx.Done(); return ctx.Err() }
	if err := c.Connect(context.Background(), profile.Profile{}); err == nil {
		t.Fatal("startup succeeded")
	}
	if c.Snapshot().State != Failed {
		t.Fatal(c.Snapshot())
	}
}

func TestSlowObserverDoesNotBlockCleanup(t *testing.T) {
	c, _, _, _, _ := harness()
	_, unsubscribe := c.Subscribe()
	defer unsubscribe()
	if err := c.Connect(context.Background(), profile.Profile{}); err != nil {
		t.Fatal(err)
	}
	if err := c.Disconnect(); err != nil {
		t.Fatal(err)
	}
}

func TestOwnerLossClosesControllerDespiteObserver(t *testing.T) {
	c, _, _, _, _ := harness()
	_, unsubscribe := c.Subscribe()
	defer unsubscribe()
	if err := c.Connect(context.Background(), profile.Profile{}); err != nil {
		t.Fatal(err)
	}
	beats := make(chan struct{})
	close(beats)
	if err := c.WatchOwner(context.Background(), beats, time.Second, func() (time.Duration, error) { return 0, nil }); err != nil {
		t.Fatal(err)
	}
	if err := c.Connect(context.Background(), profile.Profile{}); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}

func TestCleanupFailureCanBeRetried(t *testing.T) {
	c, r, n, _, _ := harness()
	if err := c.Connect(context.Background(), profile.Profile{}); err != nil {
		t.Fatal(err)
	}
	n.recoverErr = errors.New("restore failed")
	if err := c.Disconnect(); err == nil || c.Snapshot().State != Failed {
		t.Fatal("restore failure hidden")
	}
	for _, p := range r.processes {
		select {
		case <-p.Done():
		default:
			t.Fatal("engine left behind")
		}
	}
	n.recoverErr = nil
	if err := c.Disconnect(); err != nil || c.Snapshot().State != Idle {
		t.Fatal(err, c.Snapshot())
	}
}

func TestShutdownCancelsStartupAndRejectsQueuedConnect(t *testing.T) {
	c, _, _, plan, _ := harness()
	entered := make(chan struct{})
	plan.Ready = func(ctx context.Context) error { close(entered); <-ctx.Done(); return ctx.Err() }
	done := make(chan error, 1)
	go func() { done <- c.Connect(context.Background(), profile.Profile{}) }()
	<-entered
	if err := c.Shutdown(); err != nil {
		t.Fatal(err)
	}
	<-done
	if err := c.Connect(context.Background(), profile.Profile{}); !errors.Is(err, ErrClosed) {
		t.Fatal(err)
	}
}

func TestQueuedConnectCannotOvertakeDisconnect(t *testing.T) {
	c, r, _, _, _ := harness()
	c.op.Lock()
	connected := make(chan error, 1)
	disconnected := make(chan error, 1)
	go func() { connected <- c.Connect(context.Background(), profile.Profile{}) }()
	go func() { disconnected <- c.Disconnect() }()
	deadline := time.Now().Add(time.Second)
	for {
		c.mu.Lock()
		pending := c.pendingStops
		c.mu.Unlock()
		if pending > 0 {
			break
		}
		if time.Now().After(deadline) {
			c.op.Unlock()
			t.Fatal("disconnect not queued")
		}
		time.Sleep(time.Millisecond)
	}
	c.op.Unlock()
	if err := <-connected; err != nil && !errors.Is(err, ErrBusy) {
		t.Fatal(err)
	}
	if err := <-disconnected; err != nil {
		t.Fatal(err)
	}
	// Если Disconnect успел завершиться первым, более поздний Connect допустим.
	if c.Snapshot().State == Connected {
		if err := c.Disconnect(); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range r.processes {
		select {
		case <-p.Done():
		default:
			t.Fatal("session not cleaned")
		}
	}
}
