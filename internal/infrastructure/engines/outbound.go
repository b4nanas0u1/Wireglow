package engines

import (
	"bufio"
	"encoding/base64"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

type Outbound struct {
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	Server       string         `json:"server"`
	Port         int            `json:"port"`
	IP           string         `json:"ip,omitempty"`
	IPv6         string         `json:"ipv6,omitempty"`
	PrivateKey   string         `json:"private-key,omitempty"`
	PublicKey    string         `json:"public-key,omitempty"`
	PreSharedKey string         `json:"pre-shared-key,omitempty"`
	AllowedIPs   []string       `json:"allowed-ips,omitempty"`
	UDP          bool           `json:"udp"`
	MTU          int            `json:"mtu,omitempty"`
	Keepalive    int            `json:"persistent-keepalive,omitempty"`
	Amnezia      map[string]any `json:"amnezia-wg-option,omitempty"`
}

func key32(s string) bool {
	b, err := base64.StdEncoding.DecodeString(s)
	return err == nil && len(b) == 32
}

// WireGuard сохраняет один peer и все распознанные настройки; неизвестные не теряются молча.
func WireGuard(text, relay string) (Outbound, error) {
	port, err := loopbackPort(relay)
	if err != nil {
		return Outbound{}, err
	}
	out := Outbound{Name: "PROXY", Type: "wireguard", Server: "127.0.0.1", Port: port, UDP: true, MTU: 1280}
	fail := func() (Outbound, error) {
		return Outbound{}, errors.New("mihomo: invalid or unsupported WG/AWG configuration")
	}
	section := ""
	seen := map[string]bool{}
	peerCount := 0
	awg := map[string]any{}
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 4096), 256*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			switch line {
			case "[Interface]":
				if seen["interface"] {
					return fail()
				}
				seen["interface"] = true
				section = "interface"
			case "[Peer]":
				peerCount++
				if peerCount > 1 {
					return fail()
				}
				section = "peer"
			default:
				return fail()
			}
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if !ok || section == "" || value == "" || seen[section+"/"+key] {
			return fail()
		}
		seen[section+"/"+key] = true
		if section == "interface" {
			switch key {
			case "privatekey":
				if !key32(value) {
					return fail()
				}
				out.PrivateKey = value
			case "address":
				for _, address := range strings.Split(value, ",") {
					p, err := netip.ParsePrefix(strings.TrimSpace(address))
					if err != nil {
						return fail()
					}
					if p.Addr().Is4() {
						if out.IP != "" {
							return fail()
						}
						out.IP = p.Addr().String()
					} else {
						if out.IPv6 != "" {
							return fail()
						}
						out.IPv6 = p.Addr().String()
					}
				}
			case "dns": // DNS WG не превращается в системный resolver; mihomo использует отдельную policy.
				for _, address := range strings.Split(value, ",") {
					if _, err := netip.ParseAddr(strings.TrimSpace(address)); err != nil {
						return fail()
					}
				}
			case "mtu":
				n, err := strconv.Atoi(value)
				if err != nil || n < 1280 || n > 9000 {
					return fail()
				}
				out.MTU = n
			case "jc", "jmin", "jmax", "s1", "s2", "s3", "s4", "itime":
				n, err := strconv.Atoi(value)
				if err != nil || n < 0 {
					return fail()
				}
				awg[key] = n
			case "h1", "h2", "h3", "h4", "i1", "i2", "i3", "i4", "i5", "j1", "j2", "j3":
				awg[key] = value
			case "headerprotectionkey", "contentpaddingaddition", "rekeyaftertime", "rekeytimeout", "rejectaftertime", "keepalivetimeout", "maxhandshakeattempts":
				mapping := map[string]string{"headerprotectionkey": "header-protection-key", "contentpaddingaddition": "content-padding-addition", "rekeyaftertime": "rekey-after-time", "rekeytimeout": "rekey-timeout", "rejectaftertime": "reject-after-time", "keepalivetimeout": "keepalive-timeout", "maxhandshakeattempts": "max-handshake-attempts"}
				awg[mapping[key]] = value
			case "randomtrailers", "disablecookies":
				b, err := strconv.ParseBool(value)
				if err != nil {
					return fail()
				}
				if key == "randomtrailers" {
					awg["random-trailers"] = b
				} else {
					awg["disable-cookies"] = b
				}
			default:
				return fail()
			}
		} else {
			switch key {
			case "publickey":
				if !key32(value) {
					return fail()
				}
				out.PublicKey = value
			case "presharedkey":
				if !key32(value) {
					return fail()
				}
				out.PreSharedKey = value
			case "endpoint":
				h, p, err := net.SplitHostPort(value)
				n, e := strconv.Atoi(p)
				if err != nil || e != nil || h == "" || n < 1 || n > 65535 {
					return fail()
				}
			case "allowedips":
				for _, address := range strings.Split(value, ",") {
					p, err := netip.ParsePrefix(strings.TrimSpace(address))
					if err != nil {
						return fail()
					}
					out.AllowedIPs = append(out.AllowedIPs, p.Masked().String())
				}
			case "persistentkeepalive":
				n, err := strconv.Atoi(value)
				if err != nil || n < 0 || n > 65535 {
					return fail()
				}
				out.Keepalive = n
			default:
				return fail()
			}
		}
	}
	if scanner.Err() != nil || out.PrivateKey == "" || out.PublicKey == "" || out.IP == "" || peerCount != 1 || !seen["peer/endpoint"] {
		return fail()
	}
	// Для режима MATCH PROXY нужен весь IPv4 интернет, а не набор частных сетей.
	default4 := false
	for _, p := range out.AllowedIPs {
		if p == "0.0.0.0/0" {
			default4 = true
		}
	}
	if !default4 {
		return fail()
	}
	if len(awg) > 0 {
		version := 1
		for _, key := range []string{"s3", "s4", "i1", "i2", "i3", "i4", "i5"} {
			if _, ok := awg[key]; ok {
				version = 2
			}
		}
		for _, key := range []string{"h1", "h2", "h3", "h4"} {
			if v, ok := awg[key].(string); ok && strings.Contains(v, "-") {
				version = 2
			}
		}
		for key := range awg {
			if strings.Contains(key, "-") {
				version = 3
			}
		}
		legacy := false
		for _, key := range []string{"j1", "j2", "j3", "itime"} {
			if _, ok := awg[key]; ok {
				legacy = true
			}
		}
		if legacy && version == 3 {
			return fail()
		}
		if legacy {
			version = 1
		}
		awg["version"] = version
		out.Amnezia = awg
	}
	return out, nil
}
