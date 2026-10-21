package tun

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestEstablishedPeerChild(t *testing.T) {
	if os.Getenv("WIREGLOW_ESTABLISHED_PEER") != "1" {
		t.Skip("child only")
	}
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	ctx := context.Background()
	for _, args := range [][]string{{"link", "set", "lo", "up"}, {"link", "set", "peer0", "up"}, {"addr", "add", "203.0.113.99/24", "dev", "peer0"}} {
		if _, err := ip(ctx, args...); err != nil {
			t.Fatal(err)
		}
	}
	if os.Getenv("WIREGLOW_ESTABLISHED_UDP") == "1" {
		socket, err := net.ListenPacket("udp4", "203.0.113.99:444")
		if err != nil {
			t.Fatal(err)
		}
		defer socket.Close()
		io.WriteString(os.Stdout, "ready\n")
		var buffer [128]byte
		first := ""
		for {
			n, peer, err := socket.ReadFrom(buffer[:])
			if err != nil {
				return
			}
			if first == "" {
				first = peer.String()
			}
			if peer.String() != first {
				socket.WriteTo([]byte("changed"), peer)
			} else {
				socket.WriteTo(buffer[:n], peer)
			}
		}
	}
	listener, err := net.Listen("tcp4", "203.0.113.99:443")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	io.WriteString(os.Stdout, "ready\n")
	conn, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	io.Copy(conn, conn)
}
func TestEstablishedRelayChild(t *testing.T) {
	if os.Getenv("WIREGLOW_ESTABLISHED_RELAY") != "1" {
		t.Skip("child only")
	}
	network, remote := "tcp4", "203.0.113.99:443"
	if os.Getenv("WIREGLOW_ESTABLISHED_UDP") == "1" {
		network, remote = "udp4", "203.0.113.99:444"
	}
	var conn net.Conn
	var err error
	if os.Getenv("WIREGLOW_ESTABLISHED_UNCONNECTED") == "1" {
		packet, e := net.ListenPacket("udp", ":0")
		err = e
		peer, _ := net.ResolveUDPAddr("udp", remote)
		if e == nil {
			conn = &testUDPConn{PacketConn: packet, peer: peer}
		}
	} else {
		conn, err = net.DialTimeout(network, remote, time.Second)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	exchange := func() {
		conn.SetDeadline(time.Now().Add(2 * time.Second))
		if _, err := io.WriteString(conn, "alive"); err != nil {
			t.Fatal(err)
		}
		var b [5]byte
		if _, err := io.ReadFull(conn, b[:]); err != nil || string(b[:]) != "alive" {
			t.Fatal("existing relay transport failed", string(b[:]), err)
		}
	}
	exchange()
	io.WriteString(os.Stdout, "ready\n")
	bufio.NewReader(os.Stdin).ReadString('\n')
	exchange()
	io.WriteString(os.Stdout, "survived\n")
}
func TestIsolatedEstablishedRelayTCP(t *testing.T)            { testEstablishedRelay(t, false, false) }
func TestIsolatedEstablishedRelayUDP(t *testing.T)            { testEstablishedRelay(t, true, false) }
func TestIsolatedEstablishedRelayUDPUnconnected(t *testing.T) { testEstablishedRelay(t, true, true) }
func testEstablishedRelay(t *testing.T, udpMode, unconnected bool) {
	if os.Getenv("WIREGLOW_TUN_TEST") != "1" {
		t.Skip("isolated namespace required")
	}
	self, _ := os.Readlink("/proc/self/ns/net")
	host := os.Getenv("WIREGLOW_HOST_NETNS")
	if host == "" || self == host {
		t.Fatal("host namespace forbidden")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	binaryPath := os.Getenv("WIREGLOW_MIHOMO_TEST_BINARY")
	if binaryPath == "" {
		t.Fatal("mihomo required")
	}
	image, _ := os.Executable()
	peer := exec.CommandContext(ctx, "unshare", "--net", image, "-test.run=^TestEstablishedPeerChild$")
	peer.Env = append(os.Environ(), "WIREGLOW_ESTABLISHED_PEER=1", "WIREGLOW_ESTABLISHED_UDP="+strconv.Itoa(btoi(udpMode)))
	peerIn, _ := peer.StdinPipe()
	peerOut, _ := peer.StdoutPipe()
	peer.Stderr = os.Stderr
	if err := peer.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { peer.Process.Kill(); peer.Wait() }()
	for _, args := range [][]string{{"link", "set", "lo", "up"}, {"link", "add", "uplink0", "type", "veth", "peer", "name", "peer0"}, {"link", "set", "peer0", "netns", strconv.Itoa(peer.Process.Pid)}, {"link", "set", "uplink0", "up"}, {"addr", "add", "203.0.113.2/24", "dev", "uplink0"}, {"route", "add", "default", "dev", "uplink0"}} {
		if _, err := ip(ctx, args...); err != nil {
			t.Fatal(err)
		}
	}
	defer ip(context.Background(), "link", "del", "uplink0")
	io.WriteString(peerIn, "go\n")
	if line, err := bufio.NewReader(peerOut).ReadString('\n'); err != nil || line != "ready\n" {
		t.Fatal("peer readiness", line, err)
	}
	path := filepath.Join(t.TempDir(), "free-turn-proxy")
	data, _ := os.ReadFile(image)
	if err := os.WriteFile(path, data, 0700); err != nil {
		t.Fatal(err)
	}
	relay := exec.CommandContext(ctx, path, "-test.run=^TestEstablishedRelayChild$")
	relay.Env = append(os.Environ(), "WIREGLOW_ESTABLISHED_RELAY=1", "WIREGLOW_ESTABLISHED_UDP="+strconv.Itoa(btoi(udpMode)), "WIREGLOW_ESTABLISHED_UNCONNECTED="+strconv.Itoa(btoi(unconnected)))
	input, _ := relay.StdinPipe()
	output, _ := relay.StdoutPipe()
	relay.Stderr = os.Stderr
	if err := relay.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { relay.Process.Kill(); relay.Wait() }()
	reader := bufio.NewReader(output)
	if line, err := reader.ReadString('\n'); err != nil || line != "ready\n" {
		t.Fatal("relay readiness", line, err)
	}
	proxy, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()
	var tcp, udp atomic.Int32
	go fakeSOCKS(proxy, &tcp, &udp)
	server := &Server{UID: uint32(os.Getuid()), Binary: binaryPath, Runtime: t.TempDir()}
	before4, _ := ip(ctx, "-4", "-j", "rule", "show")
	before6, _ := ip(ctx, "-6", "-j", "rule", "show")
	cmd, exited, _, err := server.start(ctx, uint32(os.Getuid()), Request{ProxyPort: proxy.Addr().(*net.TCPAddr).Port, RelayPID: relay.Process.Pid})
	if err != nil {
		t.Fatal(err)
	}
	defer server.stop(cmd, exited)
	record, err := os.ReadFile(filepath.Join(server.Runtime, "journal.json"))
	if err != nil {
		t.Fatal(err)
	}
	var recovery journal
	if json.Unmarshal(record, &recovery) != nil || (!udpMode && len(recovery.RelayTCP) == 0) || (udpMode && len(recovery.RelayUDP) == 0) {
		t.Fatal("missing journaled TCP exemption")
	}
	io.WriteString(input, "go\n")
	if line, err := reader.ReadString('\n'); err != nil || line != "survived\n" {
		t.Fatal("relay transport did not survive TUN activation", line, err)
	}
	if tcp.Load() != 0 || udp.Load() != 0 {
		t.Fatal("existing transport captured by proxy")
	}
	// Cleanup after an engine crash must also remove the exact socket rules.
	cmd.Process.Kill()
	<-exited
	if err = server.stop(cmd, exited); err != nil {
		t.Fatal(err)
	}
	after4, _ := ip(ctx, "-4", "-j", "rule", "show")
	after6, _ := ip(ctx, "-6", "-j", "rule", "show")
	if string(before4) != string(after4) || string(before6) != string(after6) {
		t.Fatal("relay bypass rules leaked after engine crash")
	}

}

func btoi(value bool) int {
	if value {
		return 1
	}
	return 0
}

type testUDPConn struct {
	net.PacketConn
	peer net.Addr
}

func (c *testUDPConn) Read(b []byte) (int, error)  { n, _, err := c.ReadFrom(b); return n, err }
func (c *testUDPConn) Write(b []byte) (int, error) { return c.WriteTo(b, c.peer) }
func (c *testUDPConn) RemoteAddr() net.Addr        { return c.peer }
