package tun

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unsafe"

	winio "github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
)

func TestWindowsConfigRestrictions(t *testing.T) {
	link := windowsLink{Name: "Ethernet"}
	device := "wg-" + strings.Repeat("a", 32)
	for _, tc := range []struct {
		port               int
		path, device, peer string
	}{
		{0, `C:\bin\relay.exe`, device, ""}, {65536, `C:\bin\relay.exe`, device, ""},
		{1234, `relative.exe`, device, ""}, {1234, "C:\\bad,DIRECT\n.exe", device, ""},
		{1234, `C:\bin\relay.exe`, "foreign-device", ""}, {1234, `C:\bin\relay.exe`, device, "example.com:443"},
	} {
		if _, err := compileWindows(tc.port, tc.path, link, tc.device, tc.peer, nil); err == nil {
			t.Fatalf("unsafe config accepted: %+v", tc)
		}
	}
	b, err := compileWindows(1234, `C:\Program Files\Wireglow\bin\free-turn-proxy.exe`, link, device, "203.0.113.5:443", []string{"203.0.113.99"})
	if err != nil {
		t.Fatal(err)
	}
	var c map[string]any
	if err = json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	tun := c["tun"].(map[string]any)
	if c["mode"] != "rule" || c["find-process-mode"] != "always" || tun["auto-route"] != true || tun["strict-route"] != true {
		t.Fatal(string(b))
	}
	for _, key := range []string{"external-controller", "mixed-port", "routing-mark", "external-ui", "proxy-providers"} {
		if _, ok := c[key]; ok {
			t.Fatal("unexpected privileged capability", key)
		}
	}
	if !strings.Contains(string(b), "203.0.113.99/32") || !strings.Contains(string(b), "203.0.113.5/32") {
		t.Fatal("relay exemptions missing")
	}
	if _, ok := tun["iproute2-table-index"]; ok {
		t.Fatal("Linux routing in Windows config")
	}
	if binary := os.Getenv("WIREGLOW_MIHOMO_TEST_BINARY"); binary != "" {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		os.WriteFile(path, b, 0600)
		if out, err := exec.Command(binary, "-t", "-d", dir, "-f", path).CombinedOutput(); err != nil {
			t.Fatalf("pinned configuration rejected: %v %s", err, out)
		}
	}
}

func TestWindowsNetworkConflicts(t *testing.T) {
	physical := windowsLink{Name: "Ethernet", GUID: "physical", Kind: 6, Up: true, Gateways: []string{"192.0.2.1"}}
	for _, vpn := range []windowsLink{{Name: "Amnezia", Description: "AmneziaWG", Kind: 53, Up: true}, {Name: "Outline", Description: "Wintun Userspace Tunnel", Kind: 6, Up: true}, {Name: "Other", Kind: 131, Up: true}} {
		if _, err := selectWindowsNetwork([]windowsLink{physical, vpn}, ""); err == nil {
			t.Fatal("active tunnel accepted", vpn)
		}
		vpn.Up = false
		if _, err := selectWindowsNetwork([]windowsLink{physical, vpn}, ""); err != nil {
			t.Fatal("disconnected tunnel rejected", err)
		}
	}
	if _, err := selectWindowsNetwork([]windowsLink{physical, physical}, ""); err == nil {
		t.Fatal("ambiguous uplink accepted")
	}
	if _, err := selectWindowsNetwork(nil, ""); err == nil {
		t.Fatal("missing uplink accepted")
	}
}

func TestWindowsPipeAuthenticatesServer(t *testing.T) {
	owner, err := processOwner(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	id, _ := randomWindowsID()
	path := `\\.\pipe\Wireglow-Test-` + id
	l, err := winio.ListenPipe(path, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;" + owner.SID + ")"})
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	c := &Client{Socket: path, helper: &helperProcess{pid: uint32(os.Getpid())}}
	conn, err := c.dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	c.helper.pid++
	if conn, err = c.dial(ctx); err == nil {
		conn.Close()
		t.Fatal("pipe name squatting accepted")
	}
}

func TestWindowsReadOnlyInventory(t *testing.T) {
	if unsafe.Sizeof(windowsRoute{}) != 104 || unsafe.Offsetof(routeTable{}.Rows) != 8 {
		t.Fatal("Windows SDK route layout mismatch")
	}
	links, err := windowsLinks()
	if err != nil {
		t.Fatal(err)
	}
	if len(links) == 0 {
		t.Fatal("missing adapter inventory")
	}
	if _, err = windowsRoutes(); err != nil {
		t.Fatal(err)
	}
	if _, err = windowsTCPPeers(os.Getpid()); err != nil {
		t.Fatal(err)
	}
	// Only opens a loopback socket; does not install an adapter or change routes.
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	l.Close()
	if !windows.GetCurrentProcessToken().IsElevated() {
		f := filepath.Join(t.TempDir(), "journal.json")
		os.WriteFile(f, []byte(`{"version":1}`), 0600)
		if err := recoverWindowsJournal(context.Background(), f); err == nil {
			t.Fatal("user-writable recovery record accepted")
		}
	}
}

func TestWindowsDigestRejectsReplacement(t *testing.T) {
	p := filepath.Join(t.TempDir(), "mihomo.exe")
	os.WriteFile(p, []byte("replacement"), 0600)
	if checkWindowsEngine(p) == nil {
		t.Fatal("replacement accepted")
	}
}

func TestWindowsProtocolRejectsPrivilegedInputs(t *testing.T) {
	for _, wire := range []string{
		`{"version":1,"operation":"start","path":"C:\\Windows\\System32\\cmd.exe"}`,
		`{"version":1,"operation":"start","config":"mixed-port: 8080"}`,
		`{"version":1,"operation":"status"} {"operation":"start"}`,
		strings.Repeat("x", 4098),
	} {
		var r Request
		if err := readMessage(strings.NewReader(wire+"\n"), &r); err == nil {
			t.Fatal("unrestricted request accepted")
		}
	}
}

func TestWindowsHostVPNPreflight(t *testing.T) {
	if os.Getenv("WIREGLOW_EXPECT_ACTIVE_VPN") != "1" {
		t.Skip("read-only host VPN check opt-in required")
	}
	_, err := windowsNetwork("")
	if err == nil || !strings.Contains(err.Error(), "another VPN/tunnel is active") {
		t.Fatal("active host VPN was not detected", err)
	}
	t.Log(err) // Only a device name; no endpoints, profiles or credentials.
}

// Exercise the ShellExecuteEx ABI only from an already elevated interactive
// Sandbox user. This never auto-approves a UAC prompt or starts a network session.
func TestWindowsElevatedLaunch(t *testing.T) {
	if os.Getenv("WIREGLOW_WINDOWS_ELEVATION_TEST") != "1" {
		t.Skip("interactive Sandbox elevation smoke opt-in required")
	}
	if !windows.GetCurrentProcessToken().IsElevated() {
		t.Skip("requires an already elevated guest token; UAC consent is not automated")
	}
	requireWindowsSandbox(t)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	p, err := elevateHelper(self, []string{"-test.run=^TestWindowsElevationChild$"})
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(p.handle)
	select {
	case <-p.done:
		if p.err != nil {
			t.Fatal(p.err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("elevated child did not finish")
	}
}

func TestWindowsElevationChild(t *testing.T) {
	if os.Getenv("WIREGLOW_WINDOWS_ELEVATION_TEST") != "1" {
		t.Skip("elevation child only")
	}
}
