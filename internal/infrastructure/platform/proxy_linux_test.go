package platform

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"wireglow/internal/infrastructure/network"
)

func TestGNOMEProxyRoundTripInPrivateDBus(t *testing.T) {
	if os.Getenv("WIREGLOW_GSETTINGS_TEST") != "1" {
		t.Skip("set WIREGLOW_GSETTINGS_TEST=1 to test isolated GNOME settings")
	}
	if _, err := exec.LookPath("dbus-run-session"); err != nil {
		t.Skip("dbus-run-session unavailable")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	profile := filepath.Join(dir, "dconf-profile")
	if err := os.WriteFile(profile, []byte("user-db:wireglow_test\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "dbus-run-session", "--", executable, "-test.run=^TestGNOMEProxyHelper$")
	cmd.Env = append(os.Environ(), "WIREGLOW_GSETTINGS_HELPER="+dir, "XDG_CONFIG_HOME="+dir, "DCONF_PROFILE="+profile, "GSETTINGS_BACKEND=dconf")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("isolated GNOME: %v %s", err, output)
	}
}

func TestGNOMEProxyHelper(t *testing.T) {
	dir := os.Getenv("WIREGLOW_GSETTINGS_HELPER")
	if dir == "" {
		return
	}
	if os.Getenv("XDG_CONFIG_HOME") != dir || os.Getenv("DCONF_PROFILE") != filepath.Join(dir, "dconf-profile") {
		t.Fatal("settings test is not isolated")
	}
	settings := ProxySettings{}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	changes, err := settings.Desired("127.0.0.1:19001")
	if err != nil {
		t.Fatal(err)
	}
	before := map[string]string{}
	for _, change := range changes {
		value, err := settings.Read(ctx, change.Key)
		if err != nil {
			pair := gnomeKeys[change.Key]
			debugCtx, debugCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer debugCancel()
			output, debugErr := exec.CommandContext(debugCtx, "gsettings", "get", pair[0], pair[1]).CombinedOutput()
			t.Fatalf("%s: %v; diagnostic %v %s", change.Key, err, debugErr, output)
		}
		before[change.Key] = value
	}
	manager := &network.Manager{Path: filepath.Join(dir, "journal.json"), Settings: settings}
	if err := manager.Apply(ctx, "127.0.0.1:19001"); err != nil {
		debugCtx, debugCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer debugCancel()
		output, debugErr := exec.CommandContext(debugCtx, "gsettings", "set", "org.gnome.system.proxy.http", "host", "'127.0.0.1'").CombinedOutput()
		t.Fatalf("apply: %v; diagnostic %v %s", err, debugErr, output)
	}
	for _, change := range changes {
		value, err := settings.Read(ctx, change.Key)
		if err != nil || value != change.After {
			t.Fatal(change.Key, value, err)
		}
	}
	if err := manager.Recover(ctx); err != nil {
		t.Fatal(err)
	}
	for key, expected := range before {
		value, err := settings.Read(ctx, key)
		if err != nil || value != expected {
			t.Fatal("restore", key, value, expected, err)
		}
	}
}
