package engines

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wireglow/internal/infrastructure/engines/manifest"
	"wireglow/internal/infrastructure/persistence"
)

// Проверка схемы реальным ядром не поднимает listener/TUN и не устанавливает соединение.
func TestMihomoAcceptsGeneratedConfig(t *testing.T) {
	binary := os.Getenv("WIREGLOW_MIHOMO_TEST_BINARY")
	if binary == "" {
		t.Skip("set WIREGLOW_MIHOMO_TEST_BINARY to official " + manifest.Version("mihomo"))
	}
	for _, test := range []struct{ name, fields string }{
		{"wg", ""}, {"awg1", "Jc = 4\nJmin = 40\nJmax = 70\nS1 = 0\nS2 = 0\nH1 = 1\nH2 = 2\nH3 = 3\nH4 = 4"},
		{"awg2", "Jc = 4\nJmin = 40\nJmax = 70\nH1 = 10-20\nH2 = 30-40\nH3 = 50-60\nH4 = 70-80\nS3 = 0\nS4 = 0"},
	} {
		t.Run(test.name, func(t *testing.T) {
			text := testWG
			if test.fields != "" {
				text = strings.Replace(text, "[Peer]", test.fields+"\n[Peer]", 1)
			}
			outbound, err := WireGuard(text, "127.0.0.1:19000")
			if err != nil {
				t.Fatal(err)
			}
			checkRealConfig(t, binary, outbound)
		})
	}

}

func TestMihomoReadinessAuthenticatesRealAPI(t *testing.T) {
	binary := os.Getenv("WIREGLOW_MIHOMO_TEST_BINARY")
	if binary == "" {
		t.Skip("set WIREGLOW_MIHOMO_TEST_BINARY to official " + manifest.Version("mihomo"))
	}
	_, mixed, api, release, err := freePorts()
	if err != nil {
		t.Fatal(err)
	}
	port, _ := loopbackPort(mixed)
	outbound, err := WireGuard(testWG, "127.0.0.1:19000")
	if err != nil {
		release()
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := persistence.WriteJSON(path, config{MixedPort: port, BindAddress: "127.0.0.1", Mode: "rule", LogLevel: "silent", Controller: api, Secret: "test-token", Proxies: []Outbound{outbound}, Rules: []string{"MATCH,PROXY"}}); err != nil {
		release()
		t.Fatal(err)
	}
	release()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "-d", dir, "-f", path)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	if err := waitAPI(ctx, api, "test-token"); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: &http.Transport{Proxy: nil}, Timeout: time.Second}
	response, err := client.Get("http://" + api + "/version")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 401 {
		t.Fatal("API accessible without secret", response.StatusCode)
	}
}

func checkRealConfig(t *testing.T, binary string, outbound Outbound) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	c := config{MixedPort: 19001, BindAddress: "127.0.0.1", Mode: "rule", LogLevel: "silent", Controller: "127.0.0.1:19002", Secret: "test-token", Proxies: []Outbound{outbound}, Rules: []string{"MATCH,PROXY"}}
	if err := persistence.WriteJSON(path, c); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, binary, "-t", "-d", dir, "-f", path).CombinedOutput()
	if err != nil {
		t.Fatalf("mihomo rejected %s: %v %s", outbound.Type, err, output)
	}
}
