package tun

import "testing"

func TestDirectExemptionRejectsNonEndpointSelectors(t *testing.T) {
	for _, peer := range []string{"example.test:444", "0.0.0.0:444", "[::]:444", "[::ffff:203.0.113.7]:444", "203.0.113.7:0", "203.0.113.7/24:444", "224.0.0.1:444", "[fe80::1%eth0]:444"} {
		if _, err := directRoutes(peer, 1000); err == nil {
			t.Fatal("unsafe exemption accepted", peer)
		}
	}
	for _, peer := range []string{"", "127.0.0.1:444", "[::1]:444"} {
		routes, err := directRoutes(peer, 1000)
		if err != nil || len(routes) != 0 {
			t.Fatal("unnecessary exemption", routes, err)
		}
	}
}
