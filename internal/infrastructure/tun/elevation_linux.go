package tun

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type helperProcess struct {
	cmd        *exec.Cmd
	done       chan struct{}
	err        error
	diagnostic helperDiagnostic
}
type helperDiagnostic struct{ data []byte }

func (d *helperDiagnostic) Write(b []byte) (int, error) {
	n := len(b)
	left := 4096 - len(d.data)
	if len(b) > left {
		b = b[:left]
	}
	d.data = append(d.data, b...)
	return n, nil
}
func (p *helperProcess) result() error {
	if p.err == nil {
		return nil
	}
	var exit *exec.ExitError
	if errors.As(p.err, &exit) && (exit.ExitCode() == 126 || exit.ExitCode() == 127) {
		return errors.New("tun: elevation cancelled or denied")
	}
	detail := strings.TrimSpace(string(p.diagnostic.data))
	if detail == "" {
		return errors.New("tun: temporary helper failed to start")
	}
	return fmt.Errorf("tun: temporary helper: %s", detail)
}
func (c *Client) startTemporary(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if reason := portableUnavailable(ctx, c.PortableHelper); reason != "" {
		return errors.New(reason)
	}
	if c.helper != nil {
		select {
		case <-c.helper.done:
			c.helper = nil
		default:
			return errors.New("tun: previous helper is still stopping")
		}
	}
	cmd := exec.Command("/usr/bin/pkexec", "--disable-internal-agent", c.PortableHelper, "--portable", "--owner", strconv.Itoa(os.Getpid()))
	p := &helperProcess{cmd: cmd, done: make(chan struct{})}
	cmd.Stderr = &p.diagnostic
	if err := cmd.Start(); err != nil {
		return errors.New("tun: could not request elevation")
	}
	c.helper = p
	go func() { p.err = cmd.Wait(); close(p.done) }()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// Before authorization pkexec still belongs to the caller. After elevation
			// the root helper instead has its owner guard and bounded idle lifetime.
			_ = cmd.Process.Signal(syscall.SIGTERM)
			return ctx.Err()
		case <-p.done:
			result := p.result()
			var exit *exec.ExitError
			if errors.As(p.err, &exit) && (exit.ExitCode() == 126 || exit.ExitCode() == 127) {
				c.helper = nil
			}
			return result
		case <-ticker.C:
			conn, err := c.dial(ctx)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}
func (c *Client) waitTemporary(ctx context.Context) error {
	if c.helper == nil {
		return nil
	}
	select {
	case <-c.helper.done:
	default:
		cancelCtx, cancel := context.WithTimeout(ctx, time.Second)
		if conn, err := c.dial(cancelCtx); err == nil {
			c.sequence++
			_, _ = exchange(cancelCtx, conn, Request{Version: ProtocolVersion, ID: c.sequence, Operation: "cancel"})
			conn.Close()
		}
		cancel()
	}
	select {
	case <-c.helper.done:
		err := c.helper.result()
		udpCount := -1
		for _, line := range strings.Split(string(c.helper.diagnostic.data), "\n") {
			const udpPrefix = "wireglow-tun-udp-summary: UDP bypasses="
			if strings.HasPrefix(line, udpPrefix) {
				value, err := strconv.Atoi(strings.TrimPrefix(line, udpPrefix))
				if err == nil && value >= 0 && value <= 256 {
					udpCount = value
				}
				continue
			}

			const prefix = "wireglow-tun-summary: TCP bypasses="
			if !strings.HasPrefix(line, prefix) {
				continue
			}
			counts := strings.Split(strings.TrimPrefix(line, prefix), "; sockets present=")
			if len(counts) != 2 {
				continue
			}
			first, e1 := strconv.Atoi(counts[0])
			second, e2 := strconv.Atoi(counts[1])
			if e1 == nil && e2 == nil && first >= 0 && first <= 256 && second >= 0 && second <= first {
				c.sessionSummary = fmt.Sprintf("tun: TCP bypasses=%d; sockets present=%d", first, second)
			}
		}
		if udpCount >= 0 {
			c.sessionSummary += fmt.Sprintf("; UDP bypasses=%d", udpCount)
		}
		c.helper = nil
		return err
	case <-ctx.Done():
		return errors.New("tun: temporary helper cleanup still pending")
	}
}
