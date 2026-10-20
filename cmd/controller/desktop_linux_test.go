//go:build linux

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"wireglow/internal/infrastructure/launcher"
	"wireglow/internal/infrastructure/persistence"
	"wireglow/internal/infrastructure/platform"
	"wireglow/internal/rpc"
	pb "wireglow/internal/rpc/gen"
)

// Проверяет готовность RPC и cleanup desktop-сессии, не меняя desktop proxy.
func TestDesktopOwnerLifetime(t *testing.T) {
	if os.Getenv("WIREGLOW_SYSTEMD_TEST") != "1" {
		t.Skip("requires user systemd")
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "controller")
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	for _, crash := range []bool{false, true} {
		name := "stdin-eof"
		if crash {
			name = "owner-crash"
		}
		t.Run(name, func(t *testing.T) {
			run := filepath.Join(dir, name)
			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "desktop", "--data-dir", filepath.Join(run, "data"), "--runtime-dir", filepath.Join(run, "run"))
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err = cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer cmd.Process.Kill()
			var endpoint launcher.Endpoint
			if err = json.NewDecoder(stdout).Decode(&endpoint); err != nil {
				t.Fatal(err)
			}
			var identity platform.Identity
			if err = persistence.ReadJSON(filepath.Join(desktopRunDir(t, filepath.Join(run, "run")), "controller.json"), &identity); err != nil {
				t.Fatal(err)
			}
			conn, err := grpc.NewClient("unix://"+endpoint.Socket, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			callCtx := metadata.AppendToOutgoingContext(ctx, rpc.TokenMetadataKey, endpoint.Token)
			if _, err = pb.NewControllerServiceClient(conn).Hello(callCtx, &emptypb.Empty{}); err != nil {
				t.Fatal(err)
			}
			if crash {
				_ = cmd.Process.Kill()
			} else {
				_ = stdin.Close()
			}
			err = cmd.Wait()
			if !crash && err != nil {
				t.Fatal(err)
			}
			for {
				alive, err := identity.Alive()
				if err != nil {
					t.Fatal(err)
				}
				if !alive {
					break
				}
				select {
				case <-ctx.Done():
					t.Fatal("controller survived owner")
				case <-time.After(100 * time.Millisecond):
				}
			}
			for {
				if _, err := os.Stat(filepath.Dir(endpoint.Socket)); os.IsNotExist(err) {
					break
				} else if err != nil {
					t.Fatal(err)
				}
				select {
				case <-ctx.Done():
					t.Fatal("temporary IPC directory survived cleanup")
				case <-time.After(100 * time.Millisecond):
				}
			}
		})
	}
}
