package tun

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPortableConfigurationDoesNotRequireInstalledService(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(root, "bin", "controller")
	client := &Client{}
	if err := client.ConfigurePortable(binary); err != nil || client.PortableHelper != "" {
		t.Fatal("ordinary client changed", err)
	}
	if err := os.WriteFile(filepath.Join(root, "portable.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := client.ConfigurePortable(binary); err != nil {
		t.Fatal(err)
	}
	if client.PortableHelper != filepath.Join(root, "bin", "tun-helper") {
		t.Fatal("portable helper not selected")
	}
	if client.Socket == SocketPath || !strings.Contains(client.Socket, "portable-") {
		t.Fatal("portable reused the service endpoint")
	}
	if reason := client.Unavailable(context.Background()); strings.Contains(reason, "install") || reason == "" {
		t.Fatal("portable requires an installed service", reason)
	}
}
func TestPortableComponentSnapshotIsPrivateAndCannotFollowSymlinks(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	if err := os.WriteFile(source, []byte("original"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := copyProtected(source, target); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("changed"), 0700); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "original" {
		t.Fatal("snapshot changed", err)
	}
	info, err := os.Stat(target)
	if err != nil || info.Mode().Perm() != 0500 {
		t.Fatal("snapshot is writable", err)
	}
	if err := copyProtected(source, target); err == nil {
		t.Fatal("existing snapshot overwritten")
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(source, link); err != nil {
		t.Fatal(err)
	}
	if err := copyProtected(link, filepath.Join(root, "from-link")); err == nil {
		t.Fatal("symlink component accepted")
	}
}
func TestPortableBootstrapRejectsMissingAuthorizationBeforeWriting(t *testing.T) {
	t.Setenv("PKEXEC_UID", "")
	if err := RunPortableHelper(context.Background(), os.Getpid()); err == nil {
		t.Fatal("missing authorization accepted")
	}
}

func TestPortableCleanupRetainsOnlyValidatedSocketSummary(t *testing.T) {
	for _, line := range []string{
		"wireglow-tun-summary: TCP bypasses=12; sockets present=12\n",
		"wireglow-tun-summary: TCP bypasses=secret; sockets present=12\n",
	} {
		done := make(chan struct{})
		close(done)
		client := &Client{helper: &helperProcess{done: done, diagnostic: helperDiagnostic{data: []byte(line)}}}
		if err := client.Recover(context.Background()); err != nil {
			t.Fatal(err)
		}
		summary := client.NetworkDiagnostic()
		if strings.Contains(line, "secret") {
			if summary != "" {
				t.Fatal("unsafe summary accepted", summary)
			}
		} else if summary != "tun: TCP bypasses=12; sockets present=12" {
			t.Fatal(summary)
		}
	}
}

func TestPortableCleanupIncludesUDPSocketSummary(t *testing.T) {
	done := make(chan struct{})
	close(done)
	data := []byte("wireglow-tun-udp-summary: UDP bypasses=2\nwireglow-tun-summary: TCP bypasses=0; sockets present=0\n")
	client := &Client{helper: &helperProcess{done: done, diagnostic: helperDiagnostic{data: data}}}
	if err := client.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := client.NetworkDiagnostic(); got != "tun: TCP bypasses=0; sockets present=0; UDP bypasses=2" {
		t.Fatal(got)
	}
}
