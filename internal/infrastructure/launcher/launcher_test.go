package launcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"wireglow/internal/backend"
	"wireglow/internal/infrastructure/persistence"
	"wireglow/internal/infrastructure/platform"
)

func TestOwnerLeaseRequiresMatchingRunAndAliveIdentity(t *testing.T) {
	identity, err := platform.ProcessIdentity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		owner platform.Identity
		lease Lease
	}{
		{"matching", identity, Lease{"run", 1}},
		{"wrong run", identity, Lease{"other", 1}},
		{"wrong identity", platform.Identity{PID: identity.PID, Start: "other", Path: identity.Path}, Lease{"run", 1}},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := persistence.WriteJSON(filepath.Join(dir, "lease.json"), test.lease); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			beats := make(chan struct{}, 1)
			go feedOwner(ctx, Request{RunID: "run", Owner: test.owner}, dir, beats)
			select {
			case _, ok := <-beats:
				if ok != (test.name == "matching") {
					t.Fatal("wrong lease acceptance")
				}
			case <-time.After(time.Second):
				t.Fatal("lease checker blocked")
			}
		})
	}
}

func TestWatchdogReadyThenRecoveryAfterControllerIdentityDisappears(t *testing.T) {
	dir := t.TempDir()
	paths := platform.Paths{Data: filepath.Join(dir, "data"), Runtime: filepath.Join(dir, "runtime")}
	identity, err := platform.ProcessIdentity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	request := Request{Version: 1, RunID: "test", Owner: identity, Options: backend.Options{Paths: paths}}
	if err := persistence.WriteJSON(filepath.Join(dir, "request.json"), request); err != nil {
		t.Fatal(err)
	}
	// Тот же PID, но другой start time: watchdog не считает чужой процесс controller.
	identity.Start = "different"
	if err := persistence.WriteJSON(filepath.Join(dir, "controller.json"), identity); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := Watchdog(ctx, dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "watchdog-ready")); err != nil {
		t.Fatal(err)
	}
	var watcher platform.Identity
	if err := persistence.ReadJSON(filepath.Join(dir, "watchdog.json"), &watcher); err != nil {
		t.Fatal(err)
	}
	if watcher.PID != os.Getpid() {
		t.Fatal(watcher)
	}
}
