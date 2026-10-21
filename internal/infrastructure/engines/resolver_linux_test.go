package engines

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// Uses real mihomo, a local SOCKS server, and synthetic DNS. No TUN, system
// resolver edits, external DNS or real profile is involved.
func TestTunInnerDNSUsesProxyAndReturnsRealAddresses(t *testing.T) {
	binaryPath := os.Getenv("WIREGLOW_MIHOMO_TEST_BINARY")
	if binaryPath == "" {
		t.Skip("pinned mihomo required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	proxy, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()
	queries := make(chan string, 8)
	go func() {
		for {
			conn, err := proxy.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				conn.SetDeadline(time.Now().Add(5 * time.Second))
				var hello [3]byte
				if _, err := io.ReadFull(conn, hello[:]); err != nil {
					return
				}
				conn.Write([]byte{5, 0})
				var request [10]byte
				if _, err := io.ReadFull(conn, request[:]); err != nil {
					return
				}
				if request[1] != 1 || request[3] != 1 || net.IP(request[4:8]).String() != "203.0.113.53" || binary.BigEndian.Uint16(request[8:]) != 53 {
					queries <- "wrong DNS route"
					return
				}
				conn.Write([]byte{5, 0, 0, 1, 127, 0, 0, 1, 0, 53})
				for {
					var size [2]byte
					if _, err := io.ReadFull(conn, size[:]); err != nil {
						return
					}
					wire := make([]byte, binary.BigEndian.Uint16(size[:]))
					if _, err := io.ReadFull(conn, wire); err != nil {
						return
					}
					var message dnsmessage.Message
					if message.Unpack(wire) != nil || len(message.Questions) != 1 {
						return
					}
					queries <- message.Questions[0].Name.String()
					message.Header.Response = true
					message.Header.RecursionAvailable = true
					if message.Questions[0].Type == dnsmessage.TypeA {
						message.Answers = []dnsmessage.Resource{{Header: dnsmessage.ResourceHeader{Name: message.Questions[0].Name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET, TTL: 60}, Body: &dnsmessage.AResource{A: [4]byte{203, 0, 113, 77}}}}
					}
					wire, err = message.Pack()
					if err != nil {
						return
					}
					binary.BigEndian.PutUint16(size[:], uint16(len(wire)))
					conn.Write(append(size[:], wire...))
				}
			}()
		}
	}()
	_, mixed, api, release, err := freePorts()
	if err != nil {
		t.Fatal(err)
	}
	release()
	_, port, _ := net.SplitHostPort(proxy.Addr().String())
	socksPort, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}
	dns := tunDNS()
	// TCP DNS makes the synthetic resolver deterministic; production uses DoH
	// with the same explicit #PROXY route and redir-host policy.
	dns.Nameservers = []string{"tcp://203.0.113.53:53#PROXY"}
	mixedPort, _ := loopbackPort(mixed)
	configuration := map[string]any{"mixed-port": mixedPort, "external-controller": api, "secret": "test-token", "mode": "rule", "log-level": "silent", "dns": dns, "proxies": []any{map[string]any{"name": "PROXY", "type": "socks5", "server": "127.0.0.1", "port": socksPort}}, "rules": []string{"MATCH,PROXY"}}
	directory := t.TempDir()
	path := filepath.Join(directory, "config.json")
	data, _ := json.Marshal(configuration)
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, binaryPath, "-d", directory, "-f", path)
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cancel(); cmd.Wait() }()
	if err = waitAPI(ctx, api, "test-token"); err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://"+api+"/dns/query?name=inner.example.test&type=A", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	transport := &http.Transport{Proxy: nil}
	defer transport.CloseIdleConnections()
	client := http.Client{Transport: transport}
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, err = io.ReadAll(io.LimitReader(response.Body, 16384))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || !strings.Contains(string(data), "203.0.113.77") || strings.Contains(string(data), "198.18.") {
		t.Fatalf("inner DNS response: %s", data)
	}
	select {
	case name := <-queries:
		if name != "inner.example.test." {
			t.Fatal(name)
		}
	default:
		t.Fatal("DNS bypassed proxy")
	}
}
