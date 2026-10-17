package platform

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSingleInstanceLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "controller.lock")
	first, err := AcquireLock(path)
	if err != nil {
		t.Fatal(err)
	}
	if second, err := AcquireLock(path); !errors.Is(err, ErrLocked) {
		if second != nil {
			second.Close()
		}
		t.Fatal("second lock accepted", err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := AcquireLock(path)
	if err != nil {
		t.Fatal("lock not released", err)
	}
	second.Close()
}

func TestIdentityDetectsPIDReuse(t *testing.T) {
	id, err := ProcessIdentity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	alive, err := id.Alive()
	if err != nil || !alive {
		t.Fatal(alive, err)
	}
	id.Start += "changed"
	alive, err = id.Alive()
	if err != nil || alive {
		t.Fatal("reused identity accepted", err)
	}
}

func TestProxyAddressMustBeLoopback(t *testing.T) {
	for _, address := range []string{"0.0.0.0:9000", "192.0.2.1:9000", "127.0.0.1:0", "127.0.0.1:65536", "localhost:9000"} {
		if _, err := ProxyPort(address); err == nil {
			t.Fatal(address)
		}
	}
	if port, err := ProxyPort("127.0.0.1:9000"); err != nil || port != 9000 {
		t.Fatal(port, err)
	}
}
