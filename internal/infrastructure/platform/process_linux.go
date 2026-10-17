package platform

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"wireglow/internal/application/ports"

	"golang.org/x/sys/unix"
)

type Runner struct{}
type process struct {
	cmd  *exec.Cmd
	done chan struct{}
	err  error
	stop sync.Mutex
}

func (Runner) Start(ctx context.Context, spec ports.Command) (ports.Process, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(spec.Path) {
		return nil, errors.New("process: executable path must be absolute")
	}
	cmd := exec.Command(spec.Path, spec.Args...)
	cmd.Dir = spec.Dir
	return startCommand(ctx, cmd)
}

func startCommand(ctx context.Context, cmd *exec.Cmd) (ports.Process, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cmd.SysProcAttr = &unix.SysProcAttr{Setpgid: true}
	// Вывод ядер не публикуется: upstream пишет ключи и URL captcha.
	if err := cmd.Start(); err != nil {
		return nil, errors.New("process: start failed")
	}
	p := &process{cmd: cmd, done: make(chan struct{})}
	go func() { p.err = cmd.Wait(); close(p.done) }()
	return p, nil
}
func (p *process) Done() <-chan struct{} { return p.done }
func (p *process) Err() error {
	select {
	case <-p.done:
		return p.err
	default:
		return nil
	}
}
func (p *process) Stop(ctx context.Context) error {
	p.stop.Lock()
	defer p.stop.Unlock()
	select {
	case <-p.done:
		return nil
	default:
	}
	if err := unix.Kill(-p.cmd.Process.Pid, unix.SIGTERM); err != nil && !errors.Is(err, unix.ESRCH) {
		return err
	}
	timer := time.NewTimer(7 * time.Second)
	defer timer.Stop()
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
	case <-timer.C:
	}
	if err := unix.Kill(-p.cmd.Process.Pid, unix.SIGKILL); err != nil && !errors.Is(err, unix.ESRCH) {
		return err
	}
	select {
	case <-p.done:
		return nil
	case <-time.After(time.Second):
		return errors.New("process: kill wait timed out")
	}
}
