//go:build linux

package rpc_test

import (
	"context"
	"encoding/hex"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wireglow/internal/application/controller"
	"wireglow/internal/backend"
	"wireglow/internal/infrastructure/persistence"
	"wireglow/internal/rpc"
	pb "wireglow/internal/rpc/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Проверяет HTTP/2 gRPC на UDS без TCP listener и без изменения desktop proxy.
func TestUnixSocketTransport(t *testing.T) {
	dir, err := os.MkdirTemp("", "ft-rpc-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "control.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	if err := os.Chmod(path, 0600); err != nil {
		t.Fatal(err)
	}
	lifetime, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := controller.New(nil, nil, &network{}, controller.Options{})
	defer c.Shutdown()
	token := []byte(strings.Repeat("u", 32))
	server, err := rpc.NewServer(&backend.Backend{Controller: c, Profiles: &persistence.Profiles{Dir: filepath.Join(dir, "profiles")}}, rpc.Options{Context: lifetime, Token: token, BuildID: "uds-build", Epoch: "uds-epoch"})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	defer func() { server.Stop(); <-done }()
	conn, err := grpc.NewClient("unix://"+path, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ctx, stop := context.WithTimeout(context.Background(), 3*time.Second)
	defer stop()
	ctx = metadata.AppendToOutgoingContext(ctx, rpc.TokenMetadataKey, hex.EncodeToString(token))
	hello, err := pb.NewControllerServiceClient(conn).Hello(ctx, &emptypb.Empty{})
	if err != nil || hello.GetBuildId() != "uds-build" {
		t.Fatalf("UDS handshake: %v %v", hello, err)
	}
}
