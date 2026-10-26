package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"wireglow/internal/infrastructure/launcher"
	"wireglow/internal/infrastructure/platform"
	"wireglow/internal/rpc"
	pb "wireglow/internal/rpc/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Вызывается только из helper с изолированными dconf и D-Bus.
func testDesktopModeSwitch(t *testing.T, ctx context.Context, binary, fixture, data, runtime, probe, certificate, profileID string, before map[string]string) {
	t.Helper()
	pidFile := filepath.Join(filepath.Dir(fixture), "engine-pids")
	owner := exec.Command(binary, "desktop", "--data-dir", data, "--runtime-dir", runtime, "--relay", fixture, "--mihomo", fixture, "--probe", probe)
	owner.Env = append(os.Environ(), "SSL_CERT_FILE="+certificate)
	stdin, err := owner.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := owner.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := owner.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stdin.Close(); _ = owner.Process.Kill(); _ = owner.Wait() }()
	var endpoint launcher.Endpoint
	if err := json.NewDecoder(stdout).Decode(&endpoint); err != nil {
		t.Fatal(err)
	}
	conn, err := grpc.NewClient("unix://"+endpoint.Socket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ctx = metadata.AppendToOutgoingContext(ctx, rpc.TokenMetadataKey, endpoint.Token)
	client := pb.NewControllerServiceClient(conn)
	if _, err := client.Connect(ctx, pb.ConnectRequest_builder{Id: profileID, Mode: pb.ProxyMode_PROXY_MODE_LOCAL}.Build()); err != nil {
		t.Fatal(err)
	}
	var initial *pb.Snapshot
	for {
		initial, err = client.GetSnapshot(ctx, &emptypb.Empty{})
		if err != nil {
			t.Fatal(err)
		}
		if initial.GetState() == pb.State_STATE_CONNECTED {
			break
		}
		if initial.GetState() == pb.State_STATE_FAILED {
			t.Fatal(initial)
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(50 * time.Millisecond):
		}
	}
	pids, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.Fields(string(pids))) != 2 {
		t.Fatalf("expected two core starts: %s", pids)
	}
	identities := []platform.Identity{}
	for _, text := range strings.Fields(string(pids)) {
		pid, err := strconv.Atoi(text)
		if err != nil {
			t.Fatal(err)
		}
		identity, err := platform.ProcessIdentity(pid)
		if err != nil {
			t.Fatal(err)
		}
		identities = append(identities, identity)
	}
	settings := platform.ProxySettings{}
	verifyRestored := func() {
		for key, expected := range before {
			actual, err := settings.Read(ctx, key)
			if err != nil || actual != expected {
				t.Fatalf("restore %s: got %s want %s (%v)", key, actual, expected, err)
			}
		}
	}
	verifyRestored()
	// Несколько циклов выявляют потерю исходных значений в журнале.
	for _, mode := range []pb.ProxyMode{pb.ProxyMode_PROXY_MODE_SYSTEM, pb.ProxyMode_PROXY_MODE_LOCAL, pb.ProxyMode_PROXY_MODE_SYSTEM, pb.ProxyMode_PROXY_MODE_LOCAL} {
		result, err := client.SwitchProxyMode(ctx, pb.ProxyModeRequest_builder{Mode: mode}.Build())
		if err != nil {
			t.Fatal(err)
		}
		if result.GetState() != pb.State_STATE_CONNECTED || result.GetActiveMode() != mode || result.GetGeneration() != initial.GetGeneration() || result.GetProfileId() != initial.GetProfileId() || result.GetMixedAddress() != initial.GetMixedAddress() {
			t.Fatal("session changed", result)
		}
		if mode == pb.ProxyMode_PROXY_MODE_SYSTEM {
			desired, err := settings.Desired(initial.GetMixedAddress())
			if err != nil {
				t.Fatal(err)
			}
			for _, change := range desired {
				actual, err := settings.Read(ctx, change.Key)
				if err != nil || actual != change.After {
					t.Fatal(change.Key, actual, err)
				}
			}
		} else {
			verifyRestored()
		}
		actual, err := os.ReadFile(pidFile)
		if err != nil || string(actual) != string(pids) {
			t.Fatal("cores restarted", string(actual), err)
		}
		for _, identity := range identities {
			if alive, err := identity.Alive(); err != nil || !alive {
				t.Fatal("core stopped", identity, err)
			}
		}
	}
	if _, err := client.Disconnect(ctx, &emptypb.Empty{}); err != nil {
		t.Fatal(err)
	}
	verifyRestored()
	if _, err := os.Stat(filepath.Join(data, "proxy-journal.json")); !os.IsNotExist(err) {
		t.Fatal("journal not completed", err)
	}
}
