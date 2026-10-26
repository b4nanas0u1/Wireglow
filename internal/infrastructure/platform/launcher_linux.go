package platform

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"wireglow/internal/application/ports"
)

type systemdProcess struct {
	ports.Process
	unit string
}

func ServiceArgs(spec ports.Command, unit string) ([]string, error) {
	if !strings.HasPrefix(unit, "wireglow-") || strings.ContainsAny(unit, "/ \n\r") || len(unit) > 120 {
		return nil, errors.New("platform: invalid service name")
	}
	if !filepath.IsAbs(spec.Path) {
		return nil, errors.New("platform: absolute executable required")
	}
	args := []string{"--user", "--wait", "--collect", "--pipe", "--service-type=exec", "--unit=" + unit, "--property=Restart=no", "--property=KillMode=mixed", "--property=TimeoutStopSec=20s", "--property=SendSIGKILL=yes"}
	if spec.Dir != "" {
		args = append(args, "--working-directory="+spec.Dir)
	}
	for _, key := range []string{"DBUS_SESSION_BUS_ADDRESS", "DISPLAY", "WAYLAND_DISPLAY", "XDG_CURRENT_DESKTOP", "XDG_RUNTIME_DIR", "XDG_CONFIG_HOME", "DCONF_PROFILE", "GSETTINGS_BACKEND", "SSL_CERT_FILE"} {
		if value := os.Getenv(key); value != "" {
			args = append(args, "--setenv="+key+"="+value)
		}
	}
	args = append(args, "--", spec.Path)
	return append(args, spec.Args...), nil
}

func startService(ctx context.Context, spec ports.Command, unit string) (ports.Process, error) {
	args, err := ServiceArgs(spec, unit)
	if err != nil {
		return nil, err
	}
	path, err := exec.LookPath("systemd-run")
	if err != nil {
		return nil, errors.New("platform: systemd user manager required")
	}
	cmd := exec.Command(path, args...)
	// Manager принадлежит UID; desktop D-Bus передаётся отдельно в --setenv.
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "DBUS_SESSION_BUS_ADDRESS=") {
			cmd.Env = append(cmd.Env, value)
		}
	}
	cmd.Env = append(cmd.Env, "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/"+strconv.Itoa(os.Getuid())+"/bus")
	p, err := startCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return &systemdProcess{p, unit}, nil
}
func StartController(ctx context.Context, spec ports.Command, unit string) (ports.Process, error) {
	return startService(ctx, spec, unit)
}
func StartRecovery(ctx context.Context, spec ports.Command, unit string) (ports.Process, error) {
	return startService(ctx, spec, unit)
}
func (p *systemdProcess) Stop(ctx context.Context) error {
	stop := exec.CommandContext(ctx, "systemctl", "--user", "stop", p.unit+".service")
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "DBUS_SESSION_BUS_ADDRESS=") {
			stop.Env = append(stop.Env, value)
		}
	}
	stop.Env = append(stop.Env, "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/"+strconv.Itoa(os.Getuid())+"/bus")
	if err := stop.Run(); err != nil {
		return errors.New("platform: service stop failed")
	}
	select {
	case <-p.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RequireManagedController запрещает запуск worker вне пользовательской службы systemd.
func RequireManagedController() error {
	if os.Getenv("INVOCATION_ID") == "" {
		return errors.New("platform: controller must be launched as systemd service")
	}
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return err
	}
	if !strings.Contains(string(data), "/wireglow-controller-") {
		return errors.New("platform: controller outside managed cgroup")
	}
	return nil
}

func GracefulStop(ctx context.Context, p ports.Process) error { return p.Stop(ctx) }
