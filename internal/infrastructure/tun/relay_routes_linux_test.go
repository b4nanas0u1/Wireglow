package tun

import "testing"

func TestProcTCPAddressAndRecoveryValidation(t *testing.T) {
	for _, test := range []struct {
		encoded, ip string
		port        uint16
	}{
		{"027100CB:3039", "203.0.113.2", 12345},
		{"B80D0120000000000000000099000000:01BB", "2001:db8::99", 443},
		{"0000000000000000FFFF0000027100CB:01BB", "203.0.113.2", 443},
	} {
		address, port, err := procAddress(test.encoded)
		if err != nil || address.String() != test.ip || port != test.port {
			t.Fatal(test.encoded, address, port, err)
		}
	}
	for _, bad := range []string{"x", "0000:1", "ZZZZZZZZ:1234", "027100CB:10000"} {
		if _, _, err := procAddress(bad); err == nil {
			t.Fatal("malformed address accepted", bad)
		}
	}
	valid := relayRoute{"-4", "203.0.113.2", "203.0.113.99", 12345, 443}
	if !valid.valid() {
		t.Fatal("valid tuple rejected")
	}
	for _, invalid := range []relayRoute{
		{"-4", "127.0.0.1", "203.0.113.99", 12345, 443},
		{"-4", "0.0.0.0", "203.0.113.99", 12345, 443},
		{"-4", "203.0.113.2", "203.0.113.99", 0, 443},
		{"-6", "203.0.113.2", "203.0.113.99", 12345, 443},
		{"-4", "203.0.113.2; command", "203.0.113.99", 12345, 443},
	} {
		if invalid.valid() {
			t.Fatal("invalid recovery tuple accepted", invalid)
		}
	}
}
