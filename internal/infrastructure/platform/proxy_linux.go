package platform

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"wireglow/internal/infrastructure/network"
)

// ProxySettings работает с GNOME GSettings текущей desktop-сессии.
type ProxySettings struct{}

var gnomeKeys = map[string][2]string{
	"mode":      {"org.gnome.system.proxy", "mode"},
	"http-host": {"org.gnome.system.proxy.http", "host"}, "http-port": {"org.gnome.system.proxy.http", "port"},
	"https-host": {"org.gnome.system.proxy.https", "host"}, "https-port": {"org.gnome.system.proxy.https", "port"},
	"same-proxy":   {"org.gnome.system.proxy", "use-same-proxy"},
	"socks-host":   {"org.gnome.system.proxy.socks", "host"},
	"socks-port":   {"org.gnome.system.proxy.socks", "port"},
	"http-auth":    {"org.gnome.system.proxy.http", "use-authentication"},
	"ignore-hosts": {"org.gnome.system.proxy", "ignore-hosts"},
}

func (ProxySettings) Desired(address string) ([]network.Change, error) {
	port, err := ProxyPort(address)
	if err != nil {
		return nil, err
	}
	return []network.Change{
		{Key: "http-host", After: "'127.0.0.1'"}, {Key: "http-port", After: strconv.Itoa(port)},
		{Key: "https-host", After: "'127.0.0.1'"}, {Key: "https-port", After: strconv.Itoa(port)},
		{Key: "same-proxy", After: "false"}, {Key: "ignore-hosts", After: "['localhost', '127.0.0.0/8', '::1']"},
		{Key: "socks-host", After: "'127.0.0.1'"}, {Key: "socks-port", After: strconv.Itoa(port)},
		{Key: "http-auth", After: "false"},
		{Key: "mode", After: "'manual'"},
	}, nil
}
func gsettings(ctx context.Context, operation, key string, values ...string) (string, error) {
	pair, ok := gnomeKeys[key]
	if !ok {
		return "", errors.New("proxy: unsupported setting")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	args := append([]string{operation, pair[0], pair[1]}, values...)
	out, err := exec.CommandContext(ctx, "gsettings", args...).Output()
	if err != nil {
		return "", errors.New("proxy: GNOME " + operation + " failed for " + key)
	}
	return strings.TrimSpace(string(out)), nil
}
func (ProxySettings) Read(ctx context.Context, key string) (string, error) {
	return gsettings(ctx, "get", key)
}
func (ProxySettings) Write(ctx context.Context, key, value string) error {
	writable, err := gsettings(ctx, "writable", key)
	if err != nil {
		return err
	}
	if writable != "true" {
		return errors.New("proxy: setting locked by policy")
	}
	_, err = gsettings(ctx, "set", key, value)
	return err
}

// SystemProxyUnavailable проверяет доступность GNOME без изменения настроек.
func SystemProxyUnavailable(ctx context.Context) string {
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	if !strings.Contains(desktop, "gnome") && !strings.Contains(desktop, "unity") {
		return "System proxy requires a GNOME desktop session"
	}
	for key := range gnomeKeys {
		if _, err := gsettings(ctx, "get", key); err != nil {
			return "GNOME proxy settings are unavailable"
		}
		writable, err := gsettings(ctx, "writable", key)
		if err != nil || writable != "true" {
			return "GNOME proxy settings are locked or unavailable"
		}
	}
	return ""
}
