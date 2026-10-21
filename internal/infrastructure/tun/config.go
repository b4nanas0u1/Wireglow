package tun

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
)

// Compile creates the entire privileged config, never accepting user YAML.
// The extra mihomo only captures packets; the user engine retains WG secrets.
func Compile(port int, relayPath, uplink, device string) ([]byte, error) {
	if port < 1 || port > 65535 || !filepath.IsAbs(relayPath) || strings.ContainsAny(relayPath, ",\n\r") || uplink == "" || device == "" {
		return nil, errors.New("tun: invalid configuration")
	}
	config := map[string]any{
		"mode": "rule", "log-level": "silent", "ipv6": true,
		"find-process-mode": "always", "interface-name": uplink, "routing-mark": 20220,
		"proxies": []any{map[string]any{"name": "PROXY", "type": "socks5", "server": "127.0.0.1", "port": port, "udp": true}},
		"rules":   []string{"PROCESS-PATH," + relayPath + ",DIRECT", "MATCH,PROXY"},
		"dns":     map[string]any{"enable": true, "ipv6": false, "enhanced-mode": "fake-ip", "fake-ip-range": "198.18.0.1/16", "nameserver": []string{"https://1.0.0.1/dns-query#PROXY"}},
		"tun": map[string]any{
			"enable": true, "device": device, "stack": "mixed", "mtu": 1400,
			// Routing belongs to the helper's journal, never mihomo auto-route.
			"auto-route": false, "auto-redirect": false,
			"iproute2-table-index": 20220,
			"dns-hijack":           []string{"any:53", "tcp://any:53"},
			"inet6-address":        []string{"fdfe:dcba:9876::1/126"},
		},
	}
	return json.Marshal(config)
}
