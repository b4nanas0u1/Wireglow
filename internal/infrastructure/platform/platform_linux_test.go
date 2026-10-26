package platform

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"syscall"
	"testing"
	"time"

	"wireglow/internal/application/ports"
)

func TestServiceLifecycleProperties(t *testing.T) {
	args, err := ServiceArgs(ports.Command{Path: "/opt/wireglow test/controller", Args: []string{"worker", "--run-dir", "/tmp/a b"}}, "wireglow-controller-test")
	if err != nil {
		t.Fatal(err)
	}
	for _, property := range []string{"--property=Restart=no", "--property=KillMode=mixed", "--property=TimeoutStopSec=20s", "--service-type=exec"} {
		if !slices.Contains(args, property) {
			t.Fatal("missing", property)
		}
	}
	if !slices.Contains(args, "/opt/wireglow test/controller") || !slices.Contains(args, "/tmp/a b") {
		t.Fatal("arguments split")
	}
	if _, err := ServiceArgs(ports.Command{Path: "/bin/false"}, "evil/other.service"); err == nil {
		t.Fatal("unsafe unit name")
	}
}

func TestLockRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lock")
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if lock, err := AcquireLock(path); err == nil {
		lock.Close()
		t.Fatal("symlink lock accepted")
	}
}

func TestProcessHelper(t *testing.T) {
	if len(os.Args) < 3 || os.Args[len(os.Args)-2] != "helper-hang" {
		return
	}
	signal.Ignore(syscall.SIGTERM)
	if err := os.WriteFile(os.Args[len(os.Args)-1], []byte("ready"), 0600); err != nil {
		os.Exit(2)
	}
	for {
		time.Sleep(time.Hour)
	}
}

func TestHungEngineIsKilledWithinDeadline(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ready := filepath.Join(t.TempDir(), "ready")
	process, err := (Runner{}).Start(context.Background(), ports.Command{Path: executable, Args: []string{"-test.run=^TestProcessHelper$", "--", "helper-hang", ready}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = process.Stop(ctx)
	}()
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("helper not ready")
		}
		time.Sleep(time.Millisecond)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := process.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case <-process.Done():
	default:
		t.Fatal("process not reaped")
	}
}

func TestDefaultPaths(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config")
	t.Setenv("XDG_CONFIG_HOME", config)
	for _, runtime := range []string{filepath.Join(t.TempDir(), "runtime"), ""} {
		t.Run(runtime, func(t *testing.T) {
			t.Setenv("XDG_RUNTIME_DIR", runtime)
			paths, err := DefaultPaths()
			if err != nil {
				t.Fatal(err)
			}
			if runtime == "" {
				runtime = filepath.Join("/run/user", strconv.Itoa(os.Getuid()))
			}
			if paths.Data != filepath.Join(config, "wireglow") || paths.Runtime != filepath.Join(runtime, "wireglow") {
				t.Fatalf("unexpected default paths: %+v", paths)
			}
		})
	}
}
