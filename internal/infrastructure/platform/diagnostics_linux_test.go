package platform

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wireglow/internal/application/ports"
)

func TestEngineDiagnosticsBoundsAndSanitizes(t *testing.T) {
	d := &EngineDiagnostics{}
	d.Write([]byte("warning [TCP] dial PROXY private.example error: connection re"))
	d.Write([]byte("fused secret-token\n"))
	if got := d.Summary(); got != "outbound connection refused" {
		t.Fatal(got)
	}
	d.Write([]byte(strings.Repeat("private", 10000) + " DNS failed\n"))
	if strings.Contains(d.Summary(), "DNS") {
		t.Fatal("oversized line retained")
	}
	d.Reset()
	if d.Summary() != "" {
		t.Fatal("old diagnostics retained")
	}
	d.Write([]byte("error all DNS requests failed https://secret.example/token\n"))
	if d.Summary() != "DNS resolution failed" {
		t.Fatal(d.Summary())
	}
}
func TestRunnerCollectsOnlyOptedInEngineDiagnostics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "engine")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf 'warning dial PROXY error: connection refused secret-token\\n' >&2\n"), 0700); err != nil {
		t.Fatal(err)
	}
	process, err := (Runner{}).Start(context.Background(), ports.Command{Path: path, CaptureDiagnostics: true})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-process.Done():
	case <-time.After(time.Second):
		t.Fatal("engine did not exit")
	}
	diagnostic := process.(ports.ProcessDiagnostic)
	if diagnostic.ProcessDiagnostic() != "outbound connection refused" {
		t.Fatal(diagnostic.ProcessDiagnostic())
	}
	diagnostic.ResetProcessDiagnostic()
	if diagnostic.ProcessDiagnostic() != "" {
		t.Fatal("reset failed")
	}
}
