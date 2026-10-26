package engines

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"wireglow/internal/domain/profile"
	"wireglow/internal/freeturn"
	"wireglow/internal/infrastructure/engines/manifest"
)

const testWG = `[Interface]
PrivateKey = AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
Address = 10.13.13.3/32, fd00::3/128
DNS = 1.1.1.1
[Peer]
PublicKey = AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=
AllowedIPs = 0.0.0.0/0, ::/0
Endpoint = remote.example:51820
PersistentKeepalive = 25`

func TestWireGuardReplacesEndpointAndKeepsKeys(t *testing.T) {
	out, err := WireGuard(testWG, "127.0.0.1:19000")
	if err != nil {
		t.Fatal(err)
	}
	if out.Server != "127.0.0.1" || out.Port != 19000 || out.IP != "10.13.13.3" || out.IPv6 != "fd00::3" || out.Keepalive != 25 || out.MTU != 1280 || !out.UDP {
		t.Fatal(out)
	}
	if out.PrivateKey != strings.Repeat("A", 43)+"=" || out.PublicKey == "" || len(out.AllowedIPs) != 2 {
		t.Fatal("WG values lost")
	}
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "remote.example") {
		t.Fatal("remote endpoint retained")
	}
}

func TestAmneziaInference(t *testing.T) {
	for _, test := range []struct {
		name, fields string
		version      int
	}{
		{"v1", "Jc = 4\nH1 = 123", 1},
		{"v2", "Jc = 4\nH1 = 12-34\nS3 = 40\nI1 = <b 0x1234>", 2},
		{"legacy", "Jc = 4\nI1 = <b 0x1234>\nJ1 = 10-20\nItime = 20", 1},
		{"v3", "Jc = 4\nHeaderProtectionKey = key\nRandomTrailers = true", 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			out, err := WireGuard(strings.Replace(testWG, "[Peer]", test.fields+"\n[Peer]", 1), "127.0.0.1:9000")
			if err != nil {
				t.Fatal(err)
			}
			if out.Amnezia["version"] != test.version || out.Amnezia["jc"] != 4 {
				t.Fatal(out.Amnezia)
			}
		})
	}
}

func TestWireGuardRejectsLossyConfigurations(t *testing.T) {
	for name, text := range map[string]string{
		"hook":              strings.Replace(testWG, "[Peer]", "PostUp = secret\n[Peer]", 1),
		"multiple peers":    testWG + "\n[Peer]\nPublicKey = anything",
		"key":               strings.Replace(testWG, strings.Repeat("A", 43)+"=", "secret", 1),
		"duplicate":         strings.Replace(testWG, "[Peer]", "MTU = 1280\nMTU = 1400\n[Peer]", 1),
		"no internet route": strings.Replace(testWG, "0.0.0.0/0", "10.0.0.0/8", 1),
		"mixed AWG":         strings.Replace(testWG, "[Peer]", "HeaderProtectionKey = secret\nJ1 = 10-20\n[Peer]", 1),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := WireGuard(text, "127.0.0.1:9000")
			if err == nil {
				t.Fatal("accepted")
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("error leaks input")
			}
		})
	}
}

func TestRelayFlagsNoOriginalURIOrWG(t *testing.T) {
	c := freeturn.Connection{Provider: "vk", Peer: "peer.example:56000", VK: &freeturn.VKSettings{CallLinks: []string{"https://vk.com/call/join/one"}}, WG: &freeturn.WireGuardSettings{Config: "WG-secret"}, Obfuscation: freeturn.Obfuscation{Mode: "none"}}
	args, err := RelayArgs(c, "127.0.0.1:9000")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"-listen 127.0.0.1:9000", "-routes=false", "-mode udp", "-transport tcp", "-links https://vk.com/call/join/one"} {
		if !strings.Contains(joined, expected) {
			t.Fatal(joined)
		}
	}
	for _, forbidden := range []string{"WG-secret", "freeturn://", "-sub"} {
		if strings.Contains(joined, forbidden) {
			t.Fatal("unsafe flags")
		}
	}
	if c.RelayMode != "" || c.TurnTransport != "" {
		t.Fatal("imported settings mutated")
	}
	if _, err := RelayArgs(c, "0.0.0.0:9000"); err == nil {
		t.Fatal("public listener accepted")
	}
}

func TestRelayKCPFlagsAndValidation(t *testing.T) {
	c := freeturn.Connection{Provider: "direct", Peer: "127.0.0.1:56000", RelayMode: "tcp", KCP: &freeturn.KCPSettings{NoDelay: 1, NC: 1, Interval: 40, Resend: 2, SndWnd: 256, RcvWnd: 512, MTU: 1100, ACKNoDelay: true}}
	args, err := RelayArgs(c, "127.0.0.1:9000")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(args, " "), "-kcp-mtu 1100") || !strings.Contains(strings.Join(args, " "), "-kcp-acknodelay=true") {
		t.Fatal(args)
	}
	c.RelayMode = "udp"
	if _, err := RelayArgs(c, "127.0.0.1:9000"); err == nil {
		t.Fatal("KCP in UDP mode accepted")
	}
}

func TestReadinessRequiresAuthenticatedExpectedVersion(t *testing.T) {
	for _, version := range []string{manifest.Version("mihomo"), "v0"} {
		t.Run(version, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer token" {
					t.Error("no authentication")
				}
				_ = json.NewEncoder(w).Encode(map[string]string{"version": version})
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := waitAPI(ctx, strings.TrimPrefix(server.URL, "http://"), "token")
			if (err == nil) != (version == manifest.Version("mihomo")) {
				t.Fatal(err)
			}
		})
	}
}

func TestReadinessWaitsForLocalProxyAfterAPI(t *testing.T) {
	for _, open := range []bool{true, false} {
		t.Run(map[bool]string{true: "delayed listener", false: "cancelled"}[open], func(t *testing.T) {
			apiReached := make(chan struct{}, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]string{"version": manifest.Version("mihomo")})
				select {
				case apiReached <- struct{}{}:
				default:
				}
			}))
			defer server.Close()
			listener, err := net.Listen("tcp4", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			mixed := listener.Addr().String()
			listener.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			result := make(chan error, 1)
			go func() { result <- waitMihomo(ctx, strings.TrimPrefix(server.URL, "http://"), "token", mixed) }()
			select {
			case <-apiReached:
			case <-ctx.Done():
				t.Fatal("API was not queried")
			}
			select {
			case err := <-result:
				t.Fatalf("ready before local proxy: %v", err)
			case <-time.After(50 * time.Millisecond):
			}
			if open {
				listener, err = net.Listen("tcp4", mixed)
				if err != nil {
					t.Fatal(err)
				}
				defer listener.Close()
			} else {
				cancel()
			}
			select {
			case err := <-result:
				if (err == nil) != open {
					t.Fatalf("unexpected readiness result: %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("readiness did not finish")
			}
		})
	}
}

func TestProbeUsesProxyAndVerifiesTLS(t *testing.T) {
	requests := make(chan string, 4)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests <- r.Method + " " + r.Host; w.WriteHeader(502) }))
	defer proxy.Close()
	u, _ := url.Parse(proxy.URL)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := Probe(ctx, u.Host, "https://probe.example/test", 204); err == nil {
		t.Fatal("unconfirmed path accepted")
	}
	select {
	case request := <-requests:
		if request != "CONNECT probe.example:443" {
			t.Fatal(request)
		}
	default:
		t.Fatal("proxy bypassed")
	}
	if err := Probe(ctx, u.Host, "http://probe.example/test", 204); err == nil {
		t.Fatal("plaintext probe allowed")
	}
}

func TestDistinctSessionPorts(t *testing.T) {
	r, m, a, release, err := freePorts()
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if r == m || m == a || r == a {
		t.Fatal("colliding ports")
	}
}

func TestPreparerPortsAreStableAndOverridable(t *testing.T) {
	gotRelay, gotMixed, gotAPI, err := (Ports{}).addresses()
	if err != nil {
		t.Fatal(err)
	}
	if gotRelay != "127.0.0.1:17890" || gotMixed != "127.0.0.1:17891" || gotAPI != "127.0.0.1:17892" {
		t.Fatalf("unexpected defaults: %s, %s, %s", gotRelay, gotMixed, gotAPI)
	}
	gotRelay, gotMixed, gotAPI, err = (Ports{Relay: 29001, Mixed: 29002, API: 29003}).addresses()
	if err != nil {
		t.Fatal(err)
	}
	if gotRelay != "127.0.0.1:29001" || gotMixed != "127.0.0.1:29002" || gotAPI != "127.0.0.1:29003" {
		t.Fatalf("overrides ignored: %s, %s, %s", gotRelay, gotMixed, gotAPI)
	}
}

func TestPreparerRejectsUnsupportedOutboundBeforeSetup(t *testing.T) {
	for _, c := range []freeturn.Connection{
		{RelayMode: "tcp"},
		{RelayMode: "tcp", WG: &freeturn.WireGuardSettings{Config: testWG}},
		{RelayMode: "udp"},
		{},
	} {
		plan, err := (&Preparer{}).Prepare(context.Background(), profile.Profile{Connection: c})
		if plan != nil || err == nil || !strings.Contains(err.Error(), "WG/AWG") {
			t.Fatalf("unsupported outbound reached engine setup: plan=%v err=%v", plan, err)
		}
	}
}
