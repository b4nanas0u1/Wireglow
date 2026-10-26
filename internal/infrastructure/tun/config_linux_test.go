package tun

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrivilegedConfigRestrictions(t *testing.T) {
	for _, p := range []int{0, -1, 65536} {
		if _, err := Compile(p, "/bin/relay", "eth0", "wgtest"); err == nil {
			t.Fatal("invalid port accepted")
		}
	}
	for _, path := range []string{"relay", "/tmp/a,DIRECT\nMATCH,DIRECT"} {
		if _, err := Compile(1234, path, "eth0", "wgtest"); err == nil {
			t.Fatal("unsafe path accepted")
		}
	}
	b, err := Compile(1234, "/opt/free-turn-proxy", "eth0", "wgtest")
	if err != nil {
		t.Fatal(err)
	}
	var c map[string]any
	if err = json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	if _, ok := c["external-controller"]; ok {
		t.Fatal("privileged API exposed")
	}
	if c["mode"] != "rule" || c["routing-mark"] != float64(20220) {
		t.Fatal(string(b))
	}
	rules := c["rules"].([]any)
	if rules[0] != "PROCESS-PATH,/opt/free-turn-proxy,DIRECT" || rules[1] != "MATCH,PROXY" {
		t.Fatal(rules)
	}
	binary := os.Getenv("WIREGLOW_MIHOMO_TEST_BINARY")
	if binary == "" {
		return
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err = os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.CommandContext(context.Background(), binary, "-t", "-d", dir, "-f", path).CombinedOutput()
	if err != nil {
		t.Fatalf("real mihomo config: %v %s", err, out)
	}
}
func TestBoundedProtocol(t *testing.T) {
	var request Request
	if err := readMessage(strings.NewReader(strings.Repeat("x", 4098)), &request); err == nil {
		t.Fatal("oversized request accepted")
	}
	if err := readMessage(bytes.NewBufferString(`{"version":1,"operation":"status"}`+"\n"), &request); err != nil || request.Operation != "status" {
		t.Fatal(request, err)
	}
}

func TestPrivilegedEngineDigestRefusesReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mihomo")
	if err := os.WriteFile(path, []byte("replaced engine"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := checkEngineDigest(path); err == nil {
		t.Fatal("unapproved engine accepted")
	}
	if binary := os.Getenv("WIREGLOW_MIHOMO_TEST_BINARY"); binary != "" {
		if err := checkEngineDigest(binary); err != nil {
			t.Fatal(err)
		}
	}
}
