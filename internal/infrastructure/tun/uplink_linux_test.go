package tun

import (
	"encoding/json"
	"testing"
)

func TestActiveTunnelIgnoresDisconnectedPersistentInterfaces(t *testing.T) {
	for _, test := range []struct {
		name, link string
		active     bool
	}{
		{"outline disconnected", `{"ifname":"outline-tun0","link_type":"none","flags":["UP","NO-CARRIER"],"operstate":"DOWN"}`, false},
		{"administratively down", `{"ifname":"amn0","link_type":"none","flags":["POINTOPOINT"],"operstate":"DOWN"}`, false},
		{"down despite UP flag", `{"link_type":"none","flags":["UP"],"operstate":"DOWN"}`, false},
		{"no carrier unknown state", `{"link_type":"none","flags":["UP","NO-CARRIER"],"operstate":"UNKNOWN"}`, false},
		{"active amnezia", `{"ifname":"amn0","link_type":"none","flags":["UP","LOWER_UP"],"operstate":"UNKNOWN"}`, true},
		{"active wireguard", `{"link_type":"none","flags":["UP","LOWER_UP"],"operstate":"UNKNOWN","linkinfo":{"info_kind":"wireguard"}}`, true},
		{"active tun", `{"flags":["UP","LOWER_UP"],"operstate":"UP","linkinfo":{"info_kind":"tun"}}`, true},
		{"ethernet", `{"link_type":"ether","flags":["UP","LOWER_UP"],"operstate":"UP"}`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var link tunnelLink
			if err := json.Unmarshal([]byte(test.link), &link); err != nil {
				t.Fatal(err)
			}
			if got := link.activeTunnel(); got != test.active {
				t.Fatalf("activeTunnel=%v, want %v", got, test.active)
			}
		})
	}
}
