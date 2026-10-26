package tun

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	winio "github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

func (c *Client) ConfigurePortable(executable string) error {
	root := filepath.Dir(executable)
	if strings.EqualFold(filepath.Base(root), "bin") {
		root = filepath.Dir(root)
	}
	info, err := os.Lstat(filepath.Join(root, "portable.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("tun: invalid portable marker")
	}
	c.PortableHelper = filepath.Join(root, "bin", "tun-helper.exe")
	owner, err := processOwner(os.Getpid())
	if err != nil {
		return err
	}
	c.Socket = pipeName(owner)
	return nil
}

func pipeName(owner windowsOwner) string {
	return `\\.\pipe\Wireglow-Tun-` + strconv.Itoa(owner.Identity.PID) + "-" + owner.Identity.Start
}

func (c *Client) dial(ctx context.Context) (net.Conn, error) {
	conn, err := winio.DialPipeContext(ctx, c.Socket)
	if err != nil {
		return nil, err
	}
	// ACL alone is insufficient: another process could squat on the pipe name.
	var pid uint32
	fd, ok := conn.(interface{ Fd() uintptr })
	if !ok {
		conn.Close()
		return nil, errors.New("tun: pipe identity unavailable")
	}
	if err = windows.GetNamedPipeServerProcessId(windows.Handle(fd.Fd()), &pid); err != nil || c.helper == nil || pid != c.helper.pid {
		conn.Close()
		return nil, errors.New("tun: unexpected helper process")
	}
	return conn, nil
}

func (c *Client) Unavailable(ctx context.Context) string {
	c.mu.Lock()
	device, active, fault := c.device, c.conn != nil, c.fault
	c.mu.Unlock()
	if active {
		if fault != nil {
			return fault.Error()
		}
		if !windowsDevicePattern.MatchString(device) {
			return "tun: helper adapter identity is unavailable"
		}
		if _, err := windowsNetwork(device); err != nil {
			return err.Error()
		}
		return ""
	}
	if c.PortableHelper == "" {
		return "TUN on Windows requires the portable bundle; installed helper service is not available yet"
	}
	return portableUnavailable(ctx, c.PortableHelper)
}

func portableUnavailable(ctx context.Context, helper string) string {
	if err := ctx.Err(); err != nil {
		return "tun: capability check cancelled"
	}
	if info, err := os.Stat(helper); err != nil || !info.Mode().IsRegular() {
		return "tun: bundled tun-helper.exe is missing"
	}
	if err := checkWindowsEngine(filepath.Join(filepath.Dir(helper), "mihomo.exe")); err != nil {
		return err.Error()
	}
	if _, err := windowsNetwork(""); err != nil {
		return err.Error()
	}
	return ""
}

type helperProcess struct {
	pid    uint32
	handle windows.Handle
	done   chan struct{}
	err    error
}

func (c *Client) startTemporary(ctx context.Context) error {
	if reason := portableUnavailable(ctx, c.PortableHelper); reason != "" {
		return errors.New(reason)
	}
	if c.helper != nil {
		return errors.New("tun: previous helper cleanup is pending")
	}
	owner, err := processOwner(os.Getpid())
	if err != nil {
		return err
	}
	p, err := elevateHelper(c.PortableHelper, []string{"--portable", "--owner", strconv.Itoa(owner.Identity.PID), "--owner-start", owner.Identity.Start})
	if err != nil {
		return err
	}
	c.helper = p
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.done:
			return errors.Join(errors.New("tun: elevated helper exited before readiness; check the portable recovery journal"), p.err)
		case <-ticker.C:
			probe, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			conn, err := c.dial(probe)
			cancel()
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
		probe, cancel := context.WithTimeout(ctx, time.Second)
		if conn, err := c.dial(probe); err == nil {
			c.sequence++
			_, _ = exchange(probe, conn, Request{Version: ProtocolVersion, ID: c.sequence, Operation: "cancel"})
			conn.Close()
		}
		cancel()
	}
	select {
	case <-c.helper.done:
		err := c.helper.err
		windows.CloseHandle(c.helper.handle)
		c.helper = nil
		return err
	case <-ctx.Done():
		return errors.New("tun: elevated helper cleanup is still pending")
	}
}
