package tun

import (
	"strings"
	"testing"
)

func TestUDPRecoveryRuleIsBoundToUIDPortAndProtocol(t *testing.T) {
	route := relayUDPRoute{Family: "-4", UID: 1000, Port: 12345}
	if !route.valid() {
		t.Fatal("valid socket rejected")
	}
	args := strings.Join(route.args("del"), " ")
	if args != "-4 rule del pref 10217 uidrange 1000-1000 ipproto udp sport 12345 lookup main" {
		t.Fatal(args)
	}
	for _, invalid := range []relayUDPRoute{{Family: "-4", Port: 0}, {Family: "-x", Port: 12345}, {Family: "-4", Port: 12345, UID: ^uint32(0)}} {
		if invalid.valid() {
			t.Fatal("invalid recovery rule accepted", invalid)
		}
	}
}
