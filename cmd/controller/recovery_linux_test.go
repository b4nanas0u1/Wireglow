package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wireglow/internal/domain/profile"
	"wireglow/internal/infrastructure/launcher"
	"wireglow/internal/infrastructure/persistence"
	"wireglow/internal/infrastructure/platform"
)

func TestSystemdProxyRecoveryAfterControllerCrash(t *testing.T) { runIsolatedProxyTest(t, false) }
func TestSystemdProxyModeSwitch(t *testing.T)                   { runIsolatedProxyTest(t, true) }

func runIsolatedProxyTest(t *testing.T, switchMode bool) {
	if os.Getenv("WIREGLOW_RECOVERY_TEST") != "1" {
		t.Skip("set WIREGLOW_RECOVERY_TEST=1 for isolated systemd/GNOME recovery")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	dconfProfile := filepath.Join(dir, "dconf-profile")
	if err := os.WriteFile(dconfProfile, []byte("user-db:wireglow_recovery\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "dbus-run-session", "--", executable, "-test.run=^TestRecoveryHelper$")
	cmd.Env = append(os.Environ(), "WIREGLOW_RECOVERY_HELPER="+dir, "XDG_CONFIG_HOME="+dir, "DCONF_PROFILE="+dconfProfile, "GSETTINGS_BACKEND=dconf", "XDG_CURRENT_DESKTOP=GNOME")
	if switchMode {
		cmd.Env = append(cmd.Env, "WIREGLOW_PROXY_SWITCH=1")
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("recovery: %v %s", err, output)
	}
}

func TestRecoveryHelper(t *testing.T) {
	dir := os.Getenv("WIREGLOW_RECOVERY_HELPER")
	if dir == "" {
		return
	}
	if os.Getenv("XDG_CONFIG_HOME") != dir || os.Getenv("DCONF_PROFILE") != filepath.Join(dir, "dconf-profile") {
		t.Fatal("test not isolated")
	}
	binary, fixture := filepath.Join(dir, "controller"), filepath.Join(dir, "engine")
	for _, build := range [][]string{{"build", "-o", binary, "."}, {"build", "-o", fixture, "./testdata/engine.go"}} {
		if output, err := exec.Command("go", build...).CombinedOutput(); err != nil {
			t.Fatal(err, string(output))
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	settings := platform.ProxySettings{}
	if err := settings.Write(ctx, "mode", "'none'"); err != nil {
		t.Fatal(err)
	}
	if err := settings.Write(ctx, "http-host", "'previous.proxy'"); err != nil {
		t.Fatal(err)
	}
	changes, err := settings.Desired("127.0.0.1:19001")
	if err != nil {
		t.Fatal(err)
	}
	before := map[string]string{}
	for _, change := range changes {
		v, err := settings.Read(ctx, change.Key)
		if err != nil {
			t.Fatal(err)
		}
		before[change.Key] = v
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
	defer server.Close()
	certificate := filepath.Join(dir, "root.pem")
	if err := os.WriteFile(certificate, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	data, runtime := filepath.Join(dir, "data"), filepath.Join(dir, "runtime")
	link := filepath.Join(dir, "link")
	wg := "[Interface]\nPrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\nAddress = 10.13.13.3/32\n[Peer]\nPublicKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\nEndpoint = remote.example:51820\nAllowedIPs = 0.0.0.0/0"
	payload, err := json.Marshal(map[string]any{"v": 1, "provider": "direct", "peer": "127.0.0.1:56000", "mode": "udp", "wg": wg})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(link, []byte("freeturn://"+base64.RawURLEncoding.EncodeToString(payload)), 0600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary, "import", "--data-dir", data, "--runtime-dir", runtime, "--link-file", link).CombinedOutput()
	if err != nil {
		t.Fatal(err, string(output))
	}
	var summary profile.Summary
	if err := json.Unmarshal(output, &summary); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("WIREGLOW_PROXY_SWITCH") == "1" {
		testDesktopModeSwitch(t, ctx, binary, fixture, data, runtime, server.URL, certificate, summary.ID, before)
		return
	}
	owner := exec.Command(binary, "connect", "--data-dir", data, "--runtime-dir", runtime, "--profile", summary.ID, "--relay", fixture, "--mihomo", fixture, "--probe", server.URL, "--system-proxy")
	owner.Env = append(os.Environ(), "SSL_CERT_FILE="+certificate)
	if err := owner.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = owner.Process.Kill(); _ = owner.Wait() }()
	deadline := time.Now().Add(20 * time.Second)
	var runDir string
	connected := false
	for time.Now().Before(deadline) {
		entries, _ := os.ReadDir(runtime)
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "owner-") {
				runDir = filepath.Join(runtime, entry.Name())
			}
		}
		var report launcher.Report
		if persistence.ReadJSON(filepath.Join(runDir, "report.json"), &report) == nil {
			if report.Snapshot.State == "Connected" {
				connected = true
				break
			}
			if report.Exited {
				t.Fatal("session failed", report)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !connected {
		t.Fatal("session did not reach Connected")
	}
	if value, err := settings.Read(ctx, "mode"); err != nil || value != "'manual'" {
		t.Fatal(value, err)
	}
	var identity platform.Identity
	if err := persistence.ReadJSON(filepath.Join(runDir, "controller.json"), &identity); err != nil {
		t.Fatal(err)
	}
	controllerProcess, err := os.FindProcess(identity.PID)
	if err != nil {
		t.Fatal(err)
	}
	if err := controllerProcess.Kill(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- owner.Wait() }()
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("launcher/recovery did not finish")
	}
	for key, expected := range before {
		value, err := settings.Read(ctx, key)
		if err != nil || value != expected {
			t.Fatal("restore", key, value, expected, err)
		}
	}
	if _, err := os.Stat(filepath.Join(data, "proxy-journal.json")); !os.IsNotExist(err) {
		t.Fatal("journal not completed", err)
	}
	entries, _ := os.ReadDir(runtime)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "session-") {
			t.Fatal("secret directory left behind")
		}
	}
}
