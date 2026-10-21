package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
	"wireglow/internal/infrastructure/platform"
)

func TestProbeReportsConnectStatusWithoutPrivateProxyMessage(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(502) }))
	defer proxy.Close()
	parsed, _ := url.Parse(proxy.URL)
	err := Probe(context.Background(), parsed.Host, "https://private.example/secret", 204)
	if err == nil || err.Error() != "probe: local proxy CONNECT returned HTTP 502" {
		t.Fatal(err)
	}
}
func TestProbeDiagnosticIdentifiesPhaseAndHidesUnderlyingText(t *testing.T) {
	for _, test := range []struct {
		err         error
		phase, want string
	}{
		{io.EOF, "local proxy CONNECT", "local proxy CONNECT: peer closed"},
		{io.ErrUnexpectedEOF, "TLS handshake", "TLS handshake: peer closed"},
		{syscall.ECONNRESET, "HTTPS response", "HTTPS response: connection reset"},
		{syscall.ECONNREFUSED, "local proxy CONNECT", "local proxy CONNECT: connection refused"},
		{errors.New("private secret message"), "TLS handshake", "TLS handshake: transport error"},
	} {
		wrapped := &url.Error{Op: "Get", URL: "https://private.example/secret", Err: test.err}
		err := probeDiagnostic(context.Background(), wrapped, test.phase, 0)
		if !strings.Contains(err.Error(), test.want) || strings.Contains(err.Error(), "private") || strings.Contains(err.Error(), "secret") {
			t.Fatal(err)
		}
	}
}

func TestRealMihomoEOFRetainsSafeOutboundDiagnostic(t *testing.T) {
	binaryPath := os.Getenv("WIREGLOW_MIHOMO_TEST_BINARY")
	if binaryPath == "" {
		t.Skip("pinned mihomo required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	proxy, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()
	go func() {
		for {
			conn, err := proxy.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				conn.SetDeadline(time.Now().Add(time.Second))
				var greeting [3]byte
				if _, err := io.ReadFull(conn, greeting[:]); err != nil {
					return
				}
				conn.Write([]byte{5, 0})
				var request [10]byte
				if _, err := io.ReadFull(conn, request[:]); err != nil {
					return
				}
				conn.Write([]byte{5, 5, 0, 1, 0, 0, 0, 0, 0, 0})
			}()
		}
	}()
	_, mixed, api, release, err := freePorts()
	if err != nil {
		t.Fatal(err)
	}
	release()
	mixedPort, _ := loopbackPort(mixed)
	configuration := map[string]any{"mixed-port": mixedPort, "external-controller": api, "secret": "test-token", "mode": "rule", "log-level": "warning", "proxies": []any{map[string]any{"name": "PROXY", "type": "socks5", "server": "127.0.0.1", "port": proxy.Addr().(*net.TCPAddr).Port}}, "rules": []string{"MATCH,PROXY"}}
	directory := t.TempDir()
	path := filepath.Join(directory, "config.json")
	data, _ := json.Marshal(configuration)
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	diagnostic := &platform.EngineDiagnostics{}
	var syntheticLog bytes.Buffer
	cmd := exec.CommandContext(ctx, binaryPath, "-d", directory, "-f", path)
	cmd.Stdout = io.MultiWriter(diagnostic, &syntheticLog)
	cmd.Stderr = cmd.Stdout
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cmd.Process.Kill(); cmd.Wait() }()
	if err = waitAPI(ctx, api, "test-token"); err != nil {
		t.Fatal(err)
	}
	diagnostic.Reset()
	// Refuse the outbound TCP dial itself; a SOCKS handshake rejected after
	// lazy connection establishment is not logged by this mihomo version.
	proxy.Close()
	if err = Probe(ctx, mixed, "https://203.0.113.77/probe", 204); err == nil {
		t.Fatal("proxy failure accepted")
	}
	deadline := time.Now().Add(time.Second)
	for diagnostic.Summary() == "" && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	cmd.Process.Kill()
	cmd.Wait()
	if summary := diagnostic.Summary(); summary != "outbound dial failed" && summary != "outbound connection refused" {
		t.Fatalf("missing sanitized engine error %q, synthetic log %q", summary, syntheticLog.String())
	}

}
