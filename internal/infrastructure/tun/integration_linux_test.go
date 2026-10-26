package tun

import (
	"bufio"
	"context"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/sys/unix"
)

// This child mimics a transport process after TUN has already been enabled.
func TestTunRelayChild(t *testing.T) {
	if os.Getenv("WIREGLOW_TUN_RELAY_CHILD") != "1" {
		t.Skip("child only")
	}
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	conn, _ := net.DialTimeout("tcp4", "203.0.113.99:443", 300*time.Millisecond)
	if conn != nil {
		defer conn.Close()
	}
	udp, err := net.Dial("udp4", "203.0.113.99:444")
	if err == nil {
		_, _ = udp.Write([]byte("relay-direct"))
		defer udp.Close()
	}
	time.Sleep(200 * time.Millisecond)
	_, _ = reader.ReadString('\n')
}

// Must be explicitly run in a fresh network namespace. No host connectivity is
// needed: a synthetic SOCKS peer answers TCP and UDP and mihomo answers DNS.
func TestIsolatedTun(t *testing.T) {
	if os.Getenv("WIREGLOW_TUN_TEST") != "1" {
		t.Skip("isolated namespace opt-in required")
	}
	self, _ := os.Readlink("/proc/self/ns/net")
	host := os.Getenv("WIREGLOW_HOST_NETNS")
	if host == "" {
		host, _ = os.Readlink("/proc/1/ns/net")
	}
	if self == "" || host == "" || self == host {
		t.Fatal("refusing TUN test in host network namespace")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	links, err := ip(ctx, "-j", "link", "show")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(links), "eth") || strings.Contains(string(links), "wlan") {
		t.Fatal("namespace must start empty")
	}
	for _, args := range [][]string{{"link", "set", "lo", "up"}, {"link", "add", "uplink0", "type", "dummy"}, {"link", "set", "uplink0", "up"}, {"addr", "add", "192.0.2.2/24", "dev", "uplink0"}, {"route", "add", "default", "dev", "uplink0"}} {
		if _, err := ip(ctx, args...); err != nil {
			t.Fatal(err)
		}
	}
	defer ip(context.Background(), "link", "del", "uplink0")
	// Reproduce a persistent Outline device left UP but without an attached
	// process/carrier. It must not prevent real TUN startup or heartbeats.
	for _, args := range [][]string{{"tuntap", "add", "dev", "outline-tun0", "mode", "tun"}, {"link", "set", "outline-tun0", "up"}} {
		if _, err := ip(ctx, args...); err != nil {
			t.Fatal(err)
		}
	}
	defer ip(context.Background(), "link", "del", "outline-tun0")
	before4, _ := ip(ctx, "-4", "-j", "rule", "show")
	before6, _ := ip(ctx, "-6", "-j", "rule", "show")
	binary := os.Getenv("WIREGLOW_MIHOMO_TEST_BINARY")
	if binary == "" {
		t.Fatal("real mihomo required")
	}
	executable, _ := os.Executable()
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	relayPath := filepath.Join(t.TempDir(), "free-turn-proxy")
	if err = os.WriteFile(relayPath, data, 0700); err != nil {
		t.Fatal(err)
	}
	relay := exec.Command(relayPath, "-test.run=^TestTunRelayChild$")
	relay.Env = append(os.Environ(), "WIREGLOW_TUN_RELAY_CHILD=1")
	input, _ := relay.StdinPipe()
	if err = relay.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { relay.Process.Kill(); relay.Wait() }()
	proxy, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()
	var tcpCount, udpCount atomic.Int32
	go fakeSOCKS(proxy, &tcpCount, &udpCount)
	server := &Server{UID: uint32(os.Getuid()), Binary: binary, Runtime: t.TempDir()}
	port := proxy.Addr().(*net.TCPAddr).Port
	cmd, exited, session, err := server.start(ctx, uint32(os.Getuid()), Request{ProxyPort: port, RelayPID: relay.Process.Pid})
	if err != nil {
		t.Fatal(err)
	}
	if session == "" {
		t.Fatal("missing session ID")
	}
	defer server.stop(cmd, exited)
	t.Log("TUN started")
	conn, err := net.DialTimeout("tcp4", "203.0.113.10:80", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, _ = io.WriteString(conn, "GET / HTTP/1.1\r\nHost: example.test\r\n\r\n")
	response, err := io.ReadAll(conn)
	conn.Close()
	if err != nil || !strings.Contains(string(response), "tun-ok") || tcpCount.Load() != 1 {
		t.Fatalf("TCP capture: %q %v count=%d", response, err, tcpCount.Load())
	}

	conn6, err := net.DialTimeout("tcp6", "[2001:db8::10]:80", 5*time.Second)
	if err != nil {
		t.Fatal("IPv6 capture", err)
	}
	conn6.SetDeadline(time.Now().Add(5 * time.Second))
	_, _ = io.WriteString(conn6, "GET / HTTP/1.1\r\nHost: example.test\r\n\r\n")
	response6, err := io.ReadAll(conn6)
	conn6.Close()
	if err != nil || !strings.Contains(string(response6), "tun-ok") || tcpCount.Load() != 2 {
		t.Fatalf("IPv6 capture: %q %v count=%d", response6, err, tcpCount.Load())
	}
	udp, err := net.Dial("udp4", "203.0.113.20:2222")
	if err != nil {
		t.Fatal(err)
	}
	defer udp.Close()
	udp.SetDeadline(time.Now().Add(5 * time.Second))
	_, _ = udp.Write([]byte("tun-udp"))
	var packet [1024]byte
	n, err := udp.Read(packet[:])
	if err != nil || string(packet[:n]) != "tun-udp" || udpCount.Load() != 1 {
		t.Fatalf("UDP capture: %q %v count=%d", packet[:n], err, udpCount.Load())
	}
	dns, err := net.Dial("udp4", "203.0.113.53:53")
	if err != nil {
		t.Fatal(err)
	}
	defer dns.Close()
	dns.SetDeadline(time.Now().Add(5 * time.Second))
	query := dnsmessage.Message{Header: dnsmessage.Header{ID: 123, RecursionDesired: true}, Questions: []dnsmessage.Question{{Name: dnsmessage.MustNewName("example.test."), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}}}
	wire, _ := query.Pack()
	_, _ = dns.Write(wire)
	n, err = dns.Read(packet[:])
	if err != nil {
		t.Fatal("DNS hijack", err)
	}
	var answer dnsmessage.Message
	if err = answer.Unpack(packet[:n]); err != nil || !answer.Response || len(answer.Answers) == 0 {
		t.Fatal("DNS fake-IP response missing", answer, err)
	}
	_, _ = io.WriteString(input, "go\n")
	defer input.Close()
	time.Sleep(800 * time.Millisecond)
	if tcpCount.Load() != 2 || udpCount.Load() != 1 {
		t.Fatal("relay transport captured: loop protection failed", tcpCount.Load(), udpCount.Load())
	}
	if err = server.stop(cmd, exited); err != nil {
		t.Fatal(err)
	}
	after4, _ := ip(ctx, "-4", "-j", "rule", "show")
	after6, _ := ip(ctx, "-6", "-j", "rule", "show")
	if string(before4) != string(after4) || string(before6) != string(after6) {
		t.Fatalf("rules not restored: %s / %s", after4, after6)
	}
	if _, err = os.Stat(filepath.Join(server.Runtime, "journal.json")); !os.IsNotExist(err) {
		t.Fatal("journal retained", err)
	}
	t.Log("TCP, UDP, DNS, relay process bypass and cleanup passed")
	// A killed engine leaves policy rules; the journal must recover them.
	cmd2, exited2, _, err := server.start(ctx, uint32(os.Getuid()), Request{ProxyPort: port, RelayPID: relay.Process.Pid})
	if err != nil {
		t.Fatal(err)
	}
	defer server.stop(cmd2, exited2)
	if err = cmd2.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	<-exited2
	if err = server.Cleanup(ctx); err != nil {
		t.Fatal("engine crash recovery", err)
	}
	after4, _ = ip(ctx, "-4", "-j", "rule", "show")
	after6, _ = ip(ctx, "-6", "-j", "rule", "show")
	if string(before4) != string(after4) || string(before6) != string(after6) {
		t.Fatal("crash recovery changed baseline")
	}
	if err = server.stop(cmd2, exited2); err != nil {
		t.Fatal(err)
	}
	// Exercise real Unix IPC and owner loss, with the same root peer check as the service.
	socket := filepath.Join(server.Runtime, "control.sock")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	server.OwnerPID = os.Getpid()
	handled := make(chan bool, 4)
	go func() {
		for {
			c, e := listener.AcceptUnix()
			if e != nil {
				return
			}
			go func() { defer c.Close(); handled <- server.handle(ctx, c) }()
		}
	}()
	client := &Client{Socket: socket, RelayPID: func() int { return relay.Process.Pid }}
	if err = client.Apply(ctx, proxy.Addr().String()); err != nil {
		t.Fatal("IPC start", err)
	}
	// Dropped owner socket must stop mihomo and recover rules independently.
	client.mu.Lock()
	client.cancel()
	client.conn.Close()
	client.conn = nil
	client.mu.Unlock()
	deadline := time.Now().Add(8 * time.Second)
	for {
		_, err = os.Stat(filepath.Join(server.Runtime, "journal.json"))
		if os.IsNotExist(err) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("owner-loss recovery timed out")
		}
		time.Sleep(50 * time.Millisecond)
	}
	after4, _ = ip(ctx, "-4", "-j", "rule", "show")
	after6, _ = ip(ctx, "-6", "-j", "rule", "show")
	if string(before4) != string(after4) || string(before6) != string(after6) {
		t.Fatal("owner-loss recovery changed baseline")
	}
	select {
	case attempted := <-handled:
		if !attempted {
			t.Fatal("one-shot helper did not identify the completed owner session")
		}
	case <-time.After(time.Second):
		t.Fatal("temporary helper handler survived owner loss")
	}
	// Explicit stop also completes a one-shot helper, without a resident daemon.
	if err = client.Apply(ctx, proxy.Addr().String()); err != nil {
		t.Fatal("temporary helper restart", err)
	}
	if err = client.Recover(ctx); err != nil {
		t.Fatal("temporary helper stop", err)
	}
	select {
	case attempted := <-handled:
		if !attempted {
			t.Fatal("one-shot stop did not complete session")
		}
	case <-time.After(time.Second):
		t.Fatal("temporary helper did not finish after stop")
	}
	t.Log("engine crash, owner loss and one-shot helper stop cleanup passed")
	// A second TUN with an attached owner and carrier must still be refused.
	other, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	ifreq, err := unix.NewIfreq("amneziatest")
	if err != nil {
		t.Fatal(err)
	}
	ifreq.SetUint16(unix.IFF_TUN | unix.IFF_NO_PI)
	if err = unix.IoctlIfreq(int(other.Fd()), unix.TUNSETIFF, ifreq); err != nil {
		t.Fatal(err)
	}
	if _, err = ip(ctx, "link", "set", "amneziatest", "up"); err != nil {
		t.Fatal(err)
	}
	if _, err = server.preflight(ctx); err == nil || !strings.Contains(err.Error(), "amneziatest") {
		t.Fatal("active VPN accepted or missing interface name", err)
	}
	other.Close()
	if _, err = ip(ctx, "-4", "rule", "add", "pref", "10220", "lookup", "main"); err != nil {
		t.Fatal(err)
	}
	if _, err = server.preflight(ctx); err == nil {
		t.Fatal("foreign routing resource accepted")
	}
	if _, err = ip(ctx, "-4", "rule", "del", "pref", "10220", "lookup", "main"); err != nil {
		t.Fatal(err)
	}
	t.Log("other VPN and foreign route conflict safely refused")
}
