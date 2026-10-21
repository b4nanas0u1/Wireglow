// Package tun implements the restricted, versioned TUN helper protocol.
package tun

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strconv"
	"time"
)

const SocketPath = "/run/wireglow-tun/control.sock"
const BootstrapDNS = "1.1.1.1"
const ProtocolVersion = 1
const Lease = 15 * time.Second

// Request deliberately contains no executable, config, provider or output paths.
type Request struct {
	Version      int    `json:"version"`
	ID           uint64 `json:"id"`
	Operation    string `json:"operation"`
	Session      string `json:"session,omitempty"`
	ProxyPort    int    `json:"proxyPort,omitempty"`
	RelayPID     int    `json:"relayPID,omitempty"`
	DirectPeer   string `json:"directPeer,omitempty"`
	UDPKeepalive bool   `json:"udpKeepalive,omitempty"`
}
type Response struct {
	Version int    `json:"version"`
	ID      uint64 `json:"id"`
	Session string `json:"session,omitempty"`
	Error   string `json:"error,omitempty"`
}

func proxyPort(address string) (int, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil || host != "127.0.0.1" {
		return 0, errors.New("tun: IPv4 loopback proxy required")
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return 0, errors.New("tun: invalid proxy port")
	}
	return n, nil
}

// Each connection carries bounded newline-delimited JSON messages.
func readMessage(r io.Reader, v any) error {
	var b []byte
	for len(b) <= 4096 {
		var c [1]byte
		if _, err := io.ReadFull(r, c[:]); err != nil {
			return err
		}
		if c[0] == '\n' {
			decoder := json.NewDecoder(bytes.NewReader(b))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(v); err != nil {
				return err
			}
			var extra any
			if err := decoder.Decode(&extra); err != io.EOF {
				return errors.New("tun: trailing JSON")
			}
			return nil
		}
		b = append(b, c[0])
	}
	return errors.New("tun: message too large")
}
func exchange(ctx context.Context, conn net.Conn, request Request) (Response, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.SetDeadline(time.Now()) })
	defer stop()
	deadline := time.Now().Add(12 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return Response{}, err
	}
	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return Response{}, err
	}
	var response Response
	if err := readMessage(conn, &response); err != nil {
		return response, err
	}
	if response.Version != ProtocolVersion || response.ID != request.ID {
		return response, errors.New("tun: protocol mismatch")
	}
	if response.Error != "" {
		return response, errors.New(response.Error)
	}
	return response, nil
}
