package tun

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

// Both an explicit opt-in and the Windows Sandbox container execution service
// are required. This test must never activate TUN on the development host.
func requireWindowsSandbox(t *testing.T) {
	t.Helper()
	if os.Getenv("WIREGLOW_WINDOWS_TUN_TEST") != "1" {
		t.Skip("Windows Sandbox opt-in required")
	}
	m, err := mgr.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Disconnect()
	s, err := m.OpenService("cexecsvc")
	if err != nil {
		t.Fatal("refusing TUN outside Windows Sandbox: container service is absent", err)
	}
	defer s.Close()
	config, err := s.Config()
	if err != nil || !strings.Contains(strings.ToLower(config.BinaryPathName), "cexecsvc.exe") {
		t.Fatal("refusing TUN: invalid Sandbox service identity")
	}
	if !windows.GetCurrentProcessToken().IsElevated() {
		t.Fatal("guest administrator token required")
	}
}

// Synthetic relay, with no profile and documentation-only target addresses. The helper derives
// this process's path and SID using precisely the production authentication.
func TestWindowsRelayChild(t *testing.T) {
	if os.Getenv("WIREGLOW_WINDOWS_RELAY_CHILD") != "1" {
		t.Skip("child only")
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		if _, err := reader.ReadString('\n'); err != nil {
			return
		}
		tcp, _ := net.DialTimeout("tcp4", "203.0.113.99:443", 500*time.Millisecond)
		if tcp != nil {
			io.WriteString(tcp, "GET / HTTP/1.1\r\nHost: relay.test\r\n\r\n")
		}
		conn, err := net.Dial("udp4", "203.0.113.99:444")
		if err == nil {
			conn.Write([]byte("relay-direct"))
		}
		io.WriteString(os.Stdout, "sent\n")
		_, readErr := reader.ReadString('\n')
		if conn != nil {
			conn.Close()
		}
		if tcp != nil {
			tcp.Close()
		}
		if readErr != nil {
			return
		}
	}
}

func baselineWindowsNetwork(t *testing.T) []string {
	t.Helper()
	links, err := windowsLinks()
	if err != nil {
		t.Fatal(err)
	}
	var result []string
	for _, l := range links {
		result = append(result, l.GUID+"|"+strings.Join(l.DNS, ",")+"|"+strings.Join(l.Addresses, ","))
	}
	routes, err := windowsRoutes()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range routes {
		if r.Loopback != 0 {
			continue
		}
		result = append(result, netipRoute(r))
	}
	sort.Strings(result)
	return result
}

func netipRoute(r windowsRoute) string {
	return r.Destination.Address.ip().String() + "/" + strconvFormat(int(r.Destination.Length)) + "|" + r.NextHop.ip().String() + "|" + strconvFormat(int(r.Index)) + "|" + strconvFormat(int(r.Metric))
}
func strconvFormat(n int) string { return strconv.Itoa(n) }

func TestWindowsIsolatedTun(t *testing.T) {
	requireWindowsSandbox(t)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(filepath.Base(self), "controller.exe") {
		t.Fatal("run guest test binary as bundle/bin/controller.exe")
	}
	root := filepath.Dir(filepath.Dir(self))
	for _, p := range []string{filepath.Join(root, "data"), filepath.Join(root, "data", "runtime")} {
		if err = os.MkdirAll(p, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err = os.WriteFile(filepath.Join(root, "portable.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	// The fixture never runs elevated code from a caller-supplied path: it only
	// imitates an unprivileged relay, which the helper will never execute.
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	relayPath := filepath.Join(filepath.Dir(self), "free-turn-proxy.exe")
	if err = os.WriteFile(relayPath, data, 0700); err != nil {
		t.Fatal(err)
	}
	relay := exec.Command(relayPath, "-test.run=^TestWindowsRelayChild$")
	relay.Env = append(os.Environ(), "WIREGLOW_WINDOWS_RELAY_CHILD=1")
	input, err := relay.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := relay.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	relayOutput := bufio.NewReader(output)
	if err = relay.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { input.Close(); relay.Wait() }()
	proxy, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()
	var tcp, udp, capturedRelay atomic.Int32
	go fakeSOCKS(proxy, &tcp, &udp, func(_ byte, address []byte, port uint16) {
		if net.IP(address).String() == "203.0.113.99" && (port == 443 || port == 444) {
			capturedRelay.Add(1)
		}
	})
	before := baselineWindowsNetwork(t)
	for _, scenario := range []string{"stop", "owner-pipe-loss", "lease-timeout", "engine-crash", "helper-crash"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 65*time.Second)
			defer cancel()
			owner, err := processOwner(os.Getpid())
			if err != nil {
				t.Fatal(err)
			}
			helper := exec.Command(filepath.Join(filepath.Dir(self), "tun-helper.exe"), "--portable", "--owner", strconvFormat(os.Getpid()), "--owner-start", owner.Identity.Start)
			var diagnostics strings.Builder
			helper.Stderr = &diagnostics
			if err = helper.Start(); err != nil {
				t.Fatal(err)
			}
			h, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(helper.Process.Pid))
			if err != nil {
				t.Fatal(err)
			}
			p := &helperProcess{pid: uint32(helper.Process.Pid), handle: h, done: make(chan struct{})}
			go func() { p.err = helper.Wait(); close(p.done) }()
			c := &Client{Socket: pipeName(owner), helper: p, RelayPID: func() int { return relay.Process.Pid }}
			defer func() {
				cleanup, stop := context.WithTimeout(context.Background(), 15*time.Second)
				defer stop()
				c.Recover(cleanup)
				select {
				case <-p.done:
				default:
					helper.Process.Kill()
					<-p.done
				}
			}()
			for {
				probe, stop := context.WithTimeout(ctx, 200*time.Millisecond)
				conn, e := c.dial(probe)
				stop()
				if e == nil {
					conn.Close()
					break
				}
				select {
				case <-p.done:
					t.Fatal("helper bootstrap", p.err, diagnostics.String())
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				case <-time.After(100 * time.Millisecond):
				}
			}
			if err = c.Apply(ctx, proxy.Addr().String()); err != nil {
				t.Fatal("activation", err, diagnostics.String())
			}
			t.Log("real pinned mihomo adapter, DNS and dual-stack routes ready")
			if reason := c.Unavailable(ctx); reason != "" {
				t.Fatal("own adapter incorrectly blocks TUN capabilities", reason)
			}
			capturedRelay.Store(0)
			io.WriteString(input, "connect\n")
			defer io.WriteString(input, "release\n")
			relaySent := make(chan error, 1)
			go func() { _, err := relayOutput.ReadString('\n'); relaySent <- err }()
			select {
			case err := <-relaySent:
				if err != nil {
					t.Fatal(err)
				}
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			if capturedRelay.Load() != 0 {
				t.Fatal("relay TCP/UDP transport was forwarded into PROXY")
			}
			for _, target := range []string{"203.0.113.10:80", "[2001:db8::10]:80"} {
				conn, err := net.DialTimeout("tcp", target, 5*time.Second)
				if err != nil {
					t.Fatal("TCP", target, err)
				}
				conn.SetDeadline(time.Now().Add(5 * time.Second))
				io.WriteString(conn, "GET / HTTP/1.1\r\nHost: example.test\r\n\r\n")
				b, err := io.ReadAll(conn)
				conn.Close()
				if err != nil || !strings.Contains(string(b), "tun-ok") {
					cleanup, stop := context.WithTimeout(context.Background(), 15*time.Second)
					recoveryErr := c.Recover(cleanup)
					stop()
					t.Fatal("TCP response", target, err, string(b), "SOCKS counts", tcp.Load(), udp.Load(), "recovery", recoveryErr, "helper", diagnostics.String())
				}
			}
			for _, target := range []string{"203.0.113.20:2222", "[2001:db8::20]:2222"} {
				conn, err := net.Dial("udp", target)
				if err != nil {
					t.Fatal(err)
				}
				conn.SetDeadline(time.Now().Add(5 * time.Second))
				conn.Write([]byte("tun-udp"))
				var b [100]byte
				n, err := conn.Read(b[:])
				conn.Close()
				if err != nil || string(b[:n]) != "tun-udp" {
					t.Fatal("UDP", target, err)
				}
			}
			dns, err := net.Dial("udp4", "203.0.113.53:53")
			if err != nil {
				t.Fatal(err)
			}
			dns.SetDeadline(time.Now().Add(5 * time.Second))
			query := dnsmessage.Message{Header: dnsmessage.Header{ID: 123, RecursionDesired: true}, Questions: []dnsmessage.Question{{Name: dnsmessage.MustNewName("example.test."), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}}}
			wire, _ := query.Pack()
			dns.Write(wire)
			var packet [2048]byte
			n, err := dns.Read(packet[:])
			dns.Close()
			if err != nil {
				t.Fatal("DNS interception", err)
			}
			var answer dnsmessage.Message
			if err = answer.Unpack(packet[:n]); err != nil || !answer.Response || len(answer.Answers) == 0 {
				t.Fatal("missing fake-IP response", err)
			}
			if capturedRelay.Load() != 0 {
				t.Fatal("relay TCP/UDP transport was captured")
			}
			switch scenario {
			case "stop":
				if err = c.Recover(ctx); err != nil {
					t.Fatal(err)
				}
			case "owner-pipe-loss":
				c.mu.Lock()
				c.cancel()
				c.conn.Close()
				c.conn = nil
				c.mu.Unlock()
			case "lease-timeout":
				c.mu.Lock()
				c.cancel() // Keep the owner pipe open; only missed heartbeats may stop TUN.
				c.mu.Unlock()
			case "engine-crash":
				// Locate only the helper's protected child image, never a host engine.
				if err = killWindowsTestEngine(); err != nil {
					t.Fatal(err)
				}
			case "helper-crash":
				helper.Process.Kill()
			}
			select {
			case <-p.done:
			case <-time.After(18 * time.Second):
				t.Fatal("helper failed to terminate")
			}
			journalPath := filepath.Join(root, "data", "runtime", "tun", "journal.json")
			if err = recoverWindowsJournal(ctx, journalPath); err != nil {
				t.Fatal("recovery", err, diagnostics.String())
			}
			if after := baselineWindowsNetwork(t); !reflect.DeepEqual(before, after) {
				t.Fatalf("network baseline changed\nbefore=%v\nafter=%v", before, after)
			}
			t.Log("IPv4/IPv6 TCP+UDP, DNS and cleanup verified:", scenario)
		})
	}
}

func killWindowsTestEngine() error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if !strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), "mihomo.exe") {
			continue
		}
		h, e := windows.OpenProcess(windows.PROCESS_TERMINATE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, entry.ProcessID)
		if e != nil {
			continue
		}
		var path [32768]uint16
		size := uint32(len(path))
		e = windows.QueryFullProcessImageName(h, 0, &path[0], &size)
		if e == nil && strings.Contains(strings.ToLower(windows.UTF16ToString(path[:size])), `\wireglow-tun\`) {
			e = windows.TerminateProcess(h, 1)
			windows.CloseHandle(h)
			return e
		}
		windows.CloseHandle(h)
	}
	return errors.New("test engine was not found")
}
