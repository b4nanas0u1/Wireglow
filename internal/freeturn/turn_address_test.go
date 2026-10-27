package freeturn

import "testing"

func TestSplitTURNAddress(t *testing.T) {
	for _, tc := range []struct{ address, host, port string }{
		{"", "", ""}, {"turn.example.org", "turn.example.org", ""},
		{"turn.example.org:443", "turn.example.org", "443"},
		{"192.0.2.1", "192.0.2.1", ""}, {"192.0.2.1:3478", "192.0.2.1", "3478"},
		{"2001:db8::1", "2001:db8::1", ""}, {"[2001:db8::1]", "2001:db8::1", ""},
		{"[2001:db8::1]:443", "2001:db8::1", "443"},
	} {
		t.Run(tc.address, func(t *testing.T) {
			host, port, err := SplitTURNAddress(tc.address)
			if err != nil || host != tc.host || port != tc.port {
				t.Fatalf("got (%q, %q, %v)", host, port, err)
			}
		})
	}
	for _, address := range []string{"https://turn.example.org", "host:0", "host:65536", "host:", ":443", "a,b", "host/path", "a b", "-host", "[host]:443", "[broken]", "host:+443", " host"} {
		t.Run(address, func(t *testing.T) {
			if _, _, err := SplitTURNAddress(address); err == nil {
				t.Fatal("accepted invalid address")
			}
		})
	}
}
