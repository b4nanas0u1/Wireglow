package tun

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClientLeaseReportsHelperFailure(t *testing.T) {
	socket := filepath.Join(t.TempDir(), "helper.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, e := listener.Accept()
		if e != nil {
			return
		}
		defer conn.Close()
		var req Request
		if readMessage(conn, &req) != nil {
			return
		}
		json.NewEncoder(conn).Encode(Response{Version: 1, ID: req.ID, Session: "session"})
		if readMessage(conn, &req) != nil {
			return
		}
		json.NewEncoder(conn).Encode(Response{Version: 1, ID: req.ID, Error: "tun: engine exited"})
	}()
	client := &Client{Socket: socket, RelayPID: func() int { return os.Getpid() }}
	if err = client.Apply(context.Background(), "127.0.0.1:17891"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-client.Failed():
	case <-time.After(4 * time.Second):
		t.Fatal("helper failure not reported")
	}
	if err = client.Recover(context.Background()); err == nil {
		t.Fatal("lost helper silently recovered")
	}
}
func TestHelperRejectsWrongPeerAndUnboundedInput(t *testing.T) {
	socket := filepath.Join(t.TempDir(), "helper.sock")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	server := &Server{UID: uint32(os.Getuid() + 1)}
	go func() {
		conn, e := listener.AcceptUnix()
		if e != nil {
			return
		}
		defer conn.Close()
		server.handle(context.Background(), conn)
	}()
	client := &Client{Socket: socket}
	if reason := client.Unavailable(context.Background()); reason == "" {
		t.Fatal("wrong peer accepted")
	}
}

func TestExchangeCancellationInterruptsBlockedHelper(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { var request Request; _ = readMessage(server, &request); cancel() }()
	done := make(chan error, 1)
	go func() { _, err := exchange(ctx, client, Request{Version: 1, ID: 1, Operation: "status"}); done <- err }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancellation ignored")
		}
	case <-time.After(time.Second):
		t.Fatal("blocked helper ignored cancellation")
	}
}

func TestTemporaryHelperRejectsDifferentControllerPID(t *testing.T) {
	socket := filepath.Join(t.TempDir(), "helper.sock")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	server := &Server{UID: uint32(os.Getuid()), OwnerPID: os.Getpid() + 1, OneShot: true}
	done := make(chan bool, 1)
	go func() {
		c, e := listener.AcceptUnix()
		if e != nil {
			return
		}
		defer c.Close()
		done <- server.handle(context.Background(), c)
	}()
	client := &Client{Socket: socket}
	if reason := client.Unavailable(context.Background()); reason == "" {
		t.Fatal("wrong owner PID accepted")
	}
	if attempted := <-done; attempted {
		t.Fatal("unauthorized peer extended helper lifetime")
	}
}
