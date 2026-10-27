package rpc

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"wireglow/internal/freeturn"
	pb "wireglow/internal/rpc/gen"
)

// Advanced edits never replace server-owned keys, modes or WG configuration.
func applyAdvancedEdits(c *freeturn.Connection, edits *pb.ProfileAdvancedEdits) error {
	if edits == nil {
		return nil
	}
	switch edits.GetTurnTransport() {
	case "", freeturn.TurnTransportTCP, freeturn.TurnTransportUDP:
	default:
		return errors.New("invalid TURN transport")
	}
	dns := edits.GetDns()
	switch dns.GetMode() {
	case "", freeturn.DNSModeAuto, freeturn.DNSModePlain, freeturn.DNSModeDoH:
	default:
		return errors.New("invalid DNS mode")
	}
	servers := make([]string, 0, len(dns.GetServers()))
	for _, server := range dns.GetServers() {
		server = strings.TrimSpace(server)
		host := server
		if net.ParseIP(host) == nil {
			var port string
			var err error
			host, port, err = net.SplitHostPort(server)
			n, e := strconv.Atoi(port)
			if err != nil || e != nil || n < 1 || n > 65535 {
				return errors.New("invalid DNS server: expected IP or IP:port")
			}
		}
		if net.ParseIP(host) == nil {
			return errors.New("invalid DNS server: expected IP or IP:port")
		}
		servers = append(servers, server)
	}
	if edits.GetPacingMs() < 0 || edits.GetPacingMs() > 60000 {
		return errors.New("pacing must be between 0 and 60000 ms")
	}
	if (c.Obfuscation.Mode == "" || c.Obfuscation.Mode == freeturn.ObfuscationModeNone) && edits.GetPacingMs() != int64(c.Obfuscation.TimingMs) {
		return errors.New("pacing requires obfuscation")
	}
	if c.Provider == freeturn.ProviderVK {
		switch edits.GetVkPlatform() {
		case "", freeturn.VKPlatformDesktop, freeturn.VKPlatformMobile:
		default:
			return errors.New("invalid VK platform")
		}
		address := strings.TrimSpace(edits.GetTurnAddress())
		if _, _, err := freeturn.SplitTURNAddress(address); err != nil {
			return err
		}
		c.VK.Platform, c.VK.TurnAddress = edits.GetVkPlatform(), address
		c.VK.ManualCaptcha = edits.GetManualCaptcha()
	}
	c.TurnTransport = edits.GetTurnTransport()
	c.DNS = freeturn.DNSSettings{Mode: dns.GetMode(), Servers: servers}
	c.Obfuscation.TimingMs = int(edits.GetPacingMs())
	return nil
}
