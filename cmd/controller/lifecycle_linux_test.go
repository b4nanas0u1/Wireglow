package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"wireglow/internal/domain/profile"
	"wireglow/internal/infrastructure/platform"
)

// Тест запускает только собственные временные службы и не меняет настройки сети.
func TestSystemdOwnerCrashRemovesControllerAndEngines(t *testing.T) {
	if os.Getenv("WIREGLOW_SYSTEMD_TEST") != "1" {
		t.Skip("set WIREGLOW_SYSTEMD_TEST=1 to exercise local user systemd")
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "controller")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	data, runtime := filepath.Join(dir, "data"), filepath.Join(dir, "runtime")
	linkPath := filepath.Join(dir, "link")
	link := "freeturn://" + base64.RawURLEncoding.EncodeToString([]byte(`{"v":1,"provider":"direct","peer":"127.0.0.1:56000","wg":"[Interface]\nPrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\nAddress = 10.13.13.3/32\n[Peer]\nPublicKey = AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=\nAllowedIPs = 0.0.0.0/0\nEndpoint = 127.0.0.1:51820"}`))
	if err := os.WriteFile(linkPath, []byte(link), 0600); err != nil {
		t.Fatal(err)
	}
	importCommand := exec.Command(binary, "import", "--data-dir", data, "--runtime-dir", runtime, "--link-file", linkPath)
	output, err := importCommand.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(output))
	}
	var summary profile.Summary
	if err := json.Unmarshal(output, &summary); err != nil {
		t.Fatal(err)
	}
	pidsPath := filepath.Join(dir, "engines")
	fixture := filepath.Join(dir, "fixture")
	// Только литералы теста, без пользовательского ввода. Фикстура не открывает сеть.
	script := "#!/bin/sh\ncase \"$1\" in\n-h) echo '-provider -peer -listen -mode -transport -dns-mode -obf-profile -routes -debug -obf-timing -bond'; exit 0;;\n-t) exit 0;;\nesac\necho $$ >> '" + pidsPath + "'\nexec /bin/sleep 300\n"
	if err := os.WriteFile(fixture, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	owner := exec.Command(binary, "connect", "--data-dir", data, "--runtime-dir", runtime, "--profile", summary.ID, "--relay", fixture, "--mihomo", fixture)
	if err := owner.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = owner.Process.Kill(); _ = owner.Wait() }()
	deadline := time.Now().Add(20 * time.Second)
	var pids []int
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(pidsPath)
		if err == nil {
			lines := strings.Fields(string(data))
			if len(lines) >= 2 {
				for _, line := range lines {
					pid, err := strconv.Atoi(line)
					if err != nil {
						t.Fatal(err)
					}
					pids = append(pids, pid)
				}
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if len(pids) != 2 {
		t.Fatal("controller did not start both engines")
	}
	entries, err := os.ReadDir(runtime)
	if err != nil {
		t.Fatal(err)
	}
	var runDir string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "owner-") {
			runDir = filepath.Join(runtime, entry.Name())
		}
	}
	var controllerID platform.Identity
	file, err := os.ReadFile(filepath.Join(runDir, "controller.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(file, &controllerID); err != nil {
		t.Fatal(err)
	}
	if err := owner.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = owner.Wait()
	deadline = time.Now().Add(25 * time.Second)
	for time.Now().Before(deadline) {
		alive, err := controllerID.Alive()
		if err != nil {
			t.Fatal(err)
		}
		if !alive {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if alive, err := controllerID.Alive(); err != nil || alive {
		t.Fatal("controller survived owner", err)
	}
	for _, pid := range pids {
		if _, err := platform.ProcessIdentity(pid); err == nil {
			t.Fatal("engine survived owner:", pid)
		}
	}
}
