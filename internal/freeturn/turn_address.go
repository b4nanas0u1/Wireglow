package freeturn

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

// SplitTURNAddress accepts an optional host, host:port or [IPv6]:port.
// An omitted port preserves the provider's port; no DNS lookup is performed.
func SplitTURNAddress(address string) (host, port string, err error) {
	if address == "" {
		return "", "", nil
	}
	invalid := errors.New("freeturn: invalid TURN address")
	if strings.TrimSpace(address) != address {
		return "", "", invalid
	}
	host = address
	if strings.HasPrefix(address, "[") && strings.HasSuffix(address, "]") {
		host = address[1 : len(address)-1]
		if net.ParseIP(host) == nil {
			return "", "", invalid
		}
	} else if strings.Contains(address, ":") && net.ParseIP(address) == nil {
		host, port, err = net.SplitHostPort(address)
		if err != nil || port == "" {
			return "", "", invalid
		}
		if strings.HasPrefix(address, "[") && net.ParseIP(host) == nil {
			return "", "", invalid
		}
		n, e := strconv.Atoi(port)
		if e != nil || n < 1 || n > 65535 || strconv.Itoa(n) != port {
			return "", "", invalid
		}
	}
	if net.ParseIP(host) == nil && !validHostname(host) {
		return "", "", invalid
	}
	return host, port, nil
}

func validHostname(host string) bool {
	host = strings.TrimSuffix(host, ".")
	if len(host) == 0 || len(host) > 253 {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}
