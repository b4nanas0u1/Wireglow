package engines

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"wireglow/internal/application/ports"
	"wireglow/internal/domain/profile"
	"wireglow/internal/domain/proxymode"
	"wireglow/internal/freeturn"
	"wireglow/internal/infrastructure/tun"
)

// Prepare only runs help/schema validation, never an engine session or TUN.
func TestTunPreparationWithPinnedEngines(t *testing.T) {
	relay, mihomo := os.Getenv("WIREGLOW_RELAY_TEST_BINARY"), os.Getenv("WIREGLOW_MIHOMO_TEST_BINARY")
	if relay == "" || mihomo == "" {
		t.Skip("pinned relay and mihomo required")
	}
	original := profile.Profile{Connection: freeturn.Connection{Provider: freeturn.ProviderDirect, Peer: "localhost:56000", WG: &freeturn.WireGuardSettings{Config: testWG}, DNS: freeturn.DNSSettings{Mode: freeturn.DNSModeAuto, Servers: []string{"9.9.9.9"}}}}
	for _, mode := range []proxymode.Mode{proxymode.Local, proxymode.Tun} {
		t.Run(string(mode), func(t *testing.T) {
			preparer := &Preparer{Relay: Binary{Path: relay}, Mihomo: Binary{Path: mihomo}, RuntimeDir: t.TempDir(), ProbeURL: "https://example.test/probe"}
			plan, err := preparer.PrepareForMode(context.Background(), original, mode)
			if err != nil {
				t.Fatal(err)
			}
			defer plan.Close()
			data, err := os.ReadFile(filepath.Join(plan.Engine.Dir, "config.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			var configuration config
			if err = json.Unmarshal(data, &configuration); err != nil {
				t.Fatal(err)
			}
			if mode == proxymode.Tun {
				if configuration.DNS == nil || !configuration.DNS.Enable || configuration.DNS.Mode != "redir-host" || len(configuration.DNS.Nameservers) != 1 || configuration.DNS.Nameservers[0] != "https://1.0.0.1/dns-query#PROXY" {
					t.Fatal("inner DNS can use outer fake-IP resolver")
				}
			} else if configuration.DNS != nil {
				t.Fatal("local DNS changed")
			}
			args := strings.Join(plan.Relay.Args, " ")
			if mode == proxymode.Tun {
				if plan.DirectPeer != "127.0.0.1:56000" || !strings.Contains(args, "-peer "+plan.DirectPeer) {
					t.Fatal("direct destination not handed to helper")
				}
				bootstrapOK := strings.Contains(args, "-dns-mode plain") && strings.Contains(args, "-dns-servers "+tun.BootstrapDNS)
				if runtime.GOOS == "windows" {
					bootstrapOK = strings.Contains(args, "-dns-mode doh") && !strings.Contains(args, "-dns-servers")
				}
				if !plan.TunReady || !bootstrapOK {
					t.Fatal("bootstrap override missing")
				}
			} else if plan.DirectPeer != "" || plan.TunReady || !strings.Contains(args, "-dns-mode auto") || !strings.Contains(args, "-dns-servers 9.9.9.9") {
				t.Fatal("local settings changed")
			}
			if !plan.Relay.ClearProxyEnv {
				t.Fatal("proxy environment inherited")
			}
			if original.Connection.Peer != "localhost:56000" || original.Connection.DNS.Mode != freeturn.DNSModeAuto || original.Connection.DNS.Servers[0] != "9.9.9.9" {
				t.Fatal("stored profile mutated")
			}
		})
	}
}

func TestTunKeepaliveOnlyForVKUDPTransport(t *testing.T) {
	relay, mihomo := os.Getenv("WIREGLOW_RELAY_TEST_BINARY"), os.Getenv("WIREGLOW_MIHOMO_TEST_BINARY")
	if relay == "" || mihomo == "" {
		t.Skip("pinned engines required")
	}
	for _, tc := range []struct {
		provider, transport string
		mode                proxymode.Mode
		want                bool
	}{
		{freeturn.ProviderVK, freeturn.TurnTransportUDP, proxymode.Tun, true},
		{freeturn.ProviderVK, freeturn.TurnTransportUDP, proxymode.Local, false},
		{freeturn.ProviderVK, freeturn.TurnTransportTCP, proxymode.Tun, false},
		{freeturn.ProviderVK, "", proxymode.Tun, false},
		{freeturn.ProviderDirect, freeturn.TurnTransportUDP, proxymode.Tun, false},
	} {
		t.Run(tc.provider+"/"+tc.transport+"/"+string(tc.mode), func(t *testing.T) {
			p := profile.Profile{Connection: freeturn.Connection{Provider: tc.provider, Peer: "127.0.0.1:56000", TurnTransport: tc.transport, WG: &freeturn.WireGuardSettings{Config: testWG}}}
			if tc.provider == freeturn.ProviderVK {
				p.Connection.VK = &freeturn.VKSettings{CallLinks: []string{"https://vk.com/call/join/synthetic"}}
			}
			preparer := &Preparer{Relay: Binary{Path: relay}, Mihomo: Binary{Path: mihomo}, RuntimeDir: t.TempDir(), ProbeURL: "https://example.test/probe"}
			plan, err := preparer.PrepareForMode(t.Context(), p, tc.mode)
			if runtime.GOOS == "windows" && tc.want {
				if err == nil {
					plan.Close()
					t.Fatal("unverified TURN UDP accepted on Windows")
				}
				var diagnostic ports.DiagnosticError
				if !errors.As(err, &diagnostic) || !strings.Contains(diagnostic.PublicDiagnostic(), "select TCP TURN transport") {
					t.Fatal("TURN UDP refusal is hidden from the UI")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer plan.Close()
			if plan.RelayUDPKeepalive != tc.want {
				t.Fatal("incorrect UDP maintenance policy")
			}
		})
	}
}
