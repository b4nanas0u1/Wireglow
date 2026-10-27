// Package engines компилирует настройки и управляет контрактами внешних ядер.
package engines

import (
	"encoding/hex"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"
	"wireglow/internal/freeturn"
)

// RelayArgs передаёт только флаги: URI не может переопределить локальный listener.
func RelayArgs(c freeturn.Connection, listen string) ([]string, error) {
	if _, err := loopbackPort(listen); err != nil {
		return nil, err
	}
	if c.Provider != freeturn.ProviderVK && c.Provider != freeturn.ProviderDirect {
		return nil, errors.New("relay: unsupported provider")
	}
	host, port, err := net.SplitHostPort(c.Peer)
	p, portErr := strconv.Atoi(port)
	if err != nil || host == "" || portErr != nil || p < 1 || p > 65535 {
		return nil, errors.New("relay: invalid peer")
	}
	mode := c.RelayMode
	if mode == "" {
		mode = freeturn.RelayModeUDP
	}
	transport := c.TurnTransport
	if transport == "" {
		transport = freeturn.TurnTransportTCP
	}
	dns := c.DNS.Mode
	if dns == "" {
		dns = freeturn.DNSModeAuto
	}
	obf := c.Obfuscation.Mode
	if obf == "" {
		obf = freeturn.ObfuscationModeNone
	}
	if mode != freeturn.RelayModeUDP && mode != freeturn.RelayModeTCP || transport != freeturn.TurnTransportUDP && transport != freeturn.TurnTransportTCP {
		return nil, errors.New("relay: unsupported transport or mode")
	}
	if dns != freeturn.DNSModePlain && dns != freeturn.DNSModeDoH && dns != freeturn.DNSModeAuto {
		return nil, errors.New("relay: unsupported DNS mode")
	}
	switch obf {
	case freeturn.ObfuscationModeNone, freeturn.ObfuscationModeRTPOpus, freeturn.ObfuscationModeRTPOpus2, freeturn.ObfuscationModeRTPOpus3:
	default:
		return nil, errors.New("relay: unsupported obfuscation")
	}
	if obf != freeturn.ObfuscationModeNone {
		b, err := hex.DecodeString(c.Obfuscation.Key)
		if err != nil || len(b) != 32 {
			return nil, errors.New("relay: invalid obfuscation key")
		}
	}
	if c.Streams < 0 || c.StreamsPerCredential < 0 || c.Obfuscation.TimingMs < 0 {
		return nil, errors.New("relay: negative setting")
	}
	if mode != freeturn.RelayModeTCP && (c.KCP != nil || c.Bond) {
		return nil, errors.New("relay: KCP and bond require TCP")
	}
	args := []string{"-provider", c.Provider, "-peer", c.Peer, "-listen", listen, "-mode", mode, "-transport", transport, "-dns-mode", dns, "-obf-profile", obf, "-routes=false", "-debug=false"}
	add := func(k, v string) { args = append(args, k, v) }
	if c.ClientID != "" {
		add("-client-id", c.ClientID)
	}
	if c.Streams != 0 {
		add("-n", strconv.Itoa(c.Streams))
	}
	if c.StreamsPerCredential != 0 {
		add("-streams-per-cred", strconv.Itoa(c.StreamsPerCredential))
	}
	if len(c.DNS.Servers) > 0 {
		add("-dns-servers", strings.Join(c.DNS.Servers, ","))
	}
	if obf != freeturn.ObfuscationModeNone {
		add("-obf-key", c.Obfuscation.Key)
	}
	add("-obf-timing", (time.Duration(c.Obfuscation.TimingMs) * time.Millisecond).String())
	args = append(args, "-bond="+strconv.FormatBool(c.Bond))
	if c.Provider == freeturn.ProviderVK {
		if c.VK == nil || len(c.VK.CallLinks) == 0 {
			return nil, errors.New("relay: missing call links")
		}
		for _, link := range c.VK.CallLinks {
			if strings.TrimSpace(link) == "" || strings.Contains(link, ",") {
				return nil, errors.New("relay: invalid call link list")
			}
		}
		add("-links", strings.Join(c.VK.CallLinks, ","))
		args = append(args, "-manual-captcha="+strconv.FormatBool(c.VK.ManualCaptcha))
		if platform := c.VK.Platform; platform != "" {
			if platform != freeturn.VKPlatformDesktop && platform != freeturn.VKPlatformMobile {
				return nil, errors.New("relay: unsupported VK platform")
			}
			add("-platform", platform)
		}
		host, port, err := freeturn.SplitTURNAddress(c.VK.TurnAddress)
		if err != nil {
			return nil, err
		}
		if host != "" {
			add("-turn", host)
		}
		if port != "" {
			add("-port", port)
		}
	}
	if k := c.KCP; k != nil {
		if (k.NoDelay != 0 && k.NoDelay != 1) || (k.NC != 0 && k.NC != 1) || k.Interval <= 0 || k.Resend < 0 || k.SndWnd <= 0 || k.RcvWnd <= 0 || k.MTU < 300 || k.MTU > 1350 {
			return nil, errors.New("relay: invalid KCP settings")
		}
		for _, item := range []struct {
			k string
			v int
		}{{"nodelay", k.NoDelay}, {"interval", k.Interval}, {"resend", k.Resend}, {"nc", k.NC}, {"sndwnd", k.SndWnd}, {"rcvwnd", k.RcvWnd}, {"mtu", k.MTU}} {
			add("-kcp-"+item.k, strconv.Itoa(item.v))
		}
		args = append(args, "-kcp-acknodelay="+strconv.FormatBool(k.ACKNoDelay))
	}
	return args, nil
}

func loopbackPort(address string) (int, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil || host != "127.0.0.1" {
		return 0, errors.New("engine: endpoint must be IPv4 loopback")
	}
	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		return 0, errors.New("engine: invalid port")
	}
	return p, nil
}
