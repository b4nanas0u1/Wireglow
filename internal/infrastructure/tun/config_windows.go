package tun

import (
	"encoding/json"
	"errors"
	"net/netip"
	"path/filepath"
	"regexp"
	"strings"
)

var windowsDevicePattern = regexp.MustCompile(`^wg-[a-f0-9]{32}$`)

// These are the numeric bootstrap addresses in free-turn-proxy v4.0.3's DoH
// implementation. Its --dns-servers option does not configure DoH endpoints.
var windowsBootstrapRoutes = []string{"77.88.8.8/32", "77.88.8.1/32", "77.88.8.88/32", "77.88.8.2/32", "77.88.8.7/32", "77.88.8.3/32"}

func compileWindows(port int, relayPath string, uplink windowsLink, device, directPeer string, tcpPeers []string) ([]byte, error) {
	if port < 1 || port > 65535 || !filepath.IsAbs(relayPath) || strings.ContainsAny(relayPath, ",\n\r") || uplink.Name == "" || !windowsDevicePattern.MatchString(device) {
		return nil, errors.New("tun: invalid Windows configuration")
	}
	exclude := append([]string(nil), windowsBootstrapRoutes...)
	for _, value := range tcpPeers {
		ip, err := netip.ParseAddr(value)
		if err != nil || ip.IsUnspecified() || ip.IsMulticast() || ip.Zone() != "" {
			return nil, errors.New("tun: invalid established relay destination")
		}
		if !ip.IsLoopback() {
			exclude = append(exclude, netip.PrefixFrom(ip, ip.BitLen()).String())
		}
	}
	if directPeer != "" {
		peer, err := netip.ParseAddrPort(directPeer)
		if err != nil || peer.Port() == 0 || peer.Addr().IsUnspecified() || peer.Addr().IsMulticast() {
			return nil, errors.New("tun: invalid direct peer")
		}
		if !peer.Addr().IsLoopback() {
			exclude = append(exclude, netip.PrefixFrom(peer.Addr(), peer.Addr().BitLen()).String())
		}
	}
	config := map[string]any{
		"mode": "rule", "log-level": "silent", "ipv6": true, "find-process-mode": "always", "interface-name": uplink.Name,
		"proxies": []any{map[string]any{"name": "PROXY", "type": "socks5", "server": "127.0.0.1", "port": port, "udp": true}},
		"rules":   []string{"PROCESS-PATH," + relayPath + ",DIRECT", "MATCH,PROXY"},
		"dns":     map[string]any{"enable": true, "ipv6": false, "enhanced-mode": "fake-ip", "fake-ip-range": "198.18.0.1/16", "nameserver": []string{"https://1.0.0.1/dns-query#PROXY"}},
		"tun": map[string]any{"enable": true, "device": device, "stack": "gvisor", "mtu": 1400, "auto-route": true, "auto-detect-interface": false, "strict-route": true,
			"inet4-address": []string{"198.18.0.1/30"}, "inet6-address": []string{"fdfe:dcba:9876::1/126"}, "dns-hijack": []string{"any:53", "tcp://any:53"}, "route-exclude-address": exclude},
	}
	return json.Marshal(config)
}
