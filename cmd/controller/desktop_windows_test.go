package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/windows"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"wireglow/internal/infrastructure/launcher"
	"wireglow/internal/infrastructure/persistence"
	"wireglow/internal/infrastructure/platform"
	"wireglow/internal/rpc"
	pb "wireglow/internal/rpc/gen"
)

// A native desktop launch exercises Job Object, socket creation and RPC on
// private temporary data. It never imports a profile or changes system proxy.
func TestWindowsDesktopOwnerLifetime(t *testing.T) {
	// Keep the AF_UNIX path below the Windows sockaddr length limit.
	dir, err := os.MkdirTemp("", "ft-rpc-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	binary := filepath.Join(dir, "controller.exe")
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatal(err, string(out))
	}
	for _, crash := range []bool{false, true} {
		name := "eof"
		if crash {
			name = "crash"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()
			base := filepath.Join(dir, name)
			cmd := exec.CommandContext(ctx, binary, "desktop", "--data-dir", filepath.Join(base, "d"), "--runtime-dir", filepath.Join(base, "r"))
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
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
				t.Fatalf("desktop readiness: %v", err)
			}
			var identities [2]platform.Identity
			var handles [2]windows.Handle
			for i, file := range []string{"controller.json", "watchdog.json"} {
				if err := persistence.ReadJSON(filepath.Join(filepath.Dir(endpoint.Socket), file), &identities[i]); err != nil {
					t.Fatal(err)
				}
				handles[i], err = windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(identities[i].PID))
				if err != nil {
					t.Fatal(err)
				}
				defer windows.CloseHandle(handles[i])
			}
			conn, err := grpc.NewClient("passthrough:///desktop", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", endpoint.Socket)
			}))
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			client := pb.NewControllerServiceClient(conn)
			if _, err = client.Hello(ctx, &emptypb.Empty{}); status.Code(err) != codes.Unauthenticated {
				t.Fatalf("unauthenticated RPC accepted: %v", err)
			}
			callCtx := metadata.AppendToOutgoingContext(ctx, rpc.TokenMetadataKey, endpoint.Token)
			if _, err = client.Hello(callCtx, &emptypb.Empty{}); err != nil {
				t.Fatal(err)
			}
			if _, err = client.GetSettings(callCtx, &emptypb.Empty{}); err != nil {
				t.Fatal(err)
			}
			if _, err = client.ListProfiles(callCtx, &emptypb.Empty{}); err != nil {
				t.Fatal(err)
			}
			// Hold a snapshot of the lease across the next heartbeats. The old
			// Windows rename implementation killed the owner on its next write.
			root, err := os.OpenRoot(filepath.Dir(endpoint.Socket))
			if err != nil {
				t.Fatal(err)
			}
			leaseReader, err := root.Open("lease.json")
			root.Close()
			if err != nil {
				t.Fatal(err)
			}
			defer leaseReader.Close()
			for range 3 {
				time.Sleep(time.Second)
				if _, err := client.GetSnapshot(callCtx, &emptypb.Empty{}); err != nil {
					t.Fatalf("desktop lost RPC while replacing a lease being read: %v", err)
				}
			}
			leaseReader.Close()
			if crash {
				_ = cmd.Process.Kill()
			} else {
				_ = stdin.Close()
			}
			if err = cmd.Wait(); !crash && err != nil {
				t.Fatalf("%v: %s", err, stderr.String())
			}
			// Hold the handles from before shutdown: reopening a terminating
			// process can transiently fail with ACCESS_DENIED on Windows.
			for _, handle := range handles {
				result, err := windows.WaitForSingleObject(handle, 5000)
				if err != nil || result != windows.WAIT_OBJECT_0 {
					t.Fatalf("child survived desktop owner: %v, %v", result, err)
				}
			}
		})
	}
}
