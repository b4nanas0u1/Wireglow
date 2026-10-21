package engines

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
				if !plan.TunReady || !strings.Contains(args, "-dns-mode plain") || !strings.Contains(args, "-dns-servers "+tun.BootstrapDNS) {
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
