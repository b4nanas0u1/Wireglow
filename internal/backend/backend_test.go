package backend

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"wireglow/internal/infrastructure/network"
	"wireglow/internal/infrastructure/platform"
)

func TestSystemProxyRequiresIndependentRecovery(t *testing.T) {
	identity, err := platform.ProcessIdentity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	n := guardedNetwork{manager: &network.Manager{}, identity: identity}
	if err := n.Apply(context.Background(), "127.0.0.1:19001"); err == nil {
		t.Fatal("controller accepted itself as recovery watchdog")
	}
}

func TestBackendSingleInstanceAndClose(t *testing.T) {
	dir := t.TempDir()
	options := Options{Paths: platform.Paths{Data: filepath.Join(dir, "data"), Runtime: filepath.Join(dir, "runtime")}}
	first, err := Open(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	if second, err := Open(context.Background(), options); err == nil {
		second.Close()
		t.Fatal("second backend accepted")
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
}
