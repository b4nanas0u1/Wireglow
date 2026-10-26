package engines

import (
	"context"
	"errors"
	"net/netip"
	"testing"
)

func TestDirectPeerResolutionIsPinnedAndSanitized(t *testing.T) {
	calls := 0
	lookup := func(ctx context.Context, network, host string) ([]netip.Addr, error) {
		calls++
		if network != "ip" || host != "peer.example" {
			t.Fatal("unexpected DNS query")
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("unbounded resolution")
		}
		return []netip.Addr{netip.MustParseAddr("2001:db8::1"), netip.MustParseAddr("203.0.113.7")}, nil
	}
	for _, tc := range []struct{ input, want string }{
		{"peer.example:444", "203.0.113.7:444"},
		{"[2001:db8::1]:444", "[2001:db8::1]:444"},
		{"[::ffff:203.0.113.7]:444", "203.0.113.7:444"},
	} {
		got, err := pinDirectPeer(t.Context(), tc.input, lookup)
		if err != nil || got != tc.want {
			t.Fatal(got, err)
		}
	}
	if calls != 1 {
		t.Fatal("numeric peer unexpectedly resolved")
	}
	for _, input := range []string{"peer.example:0", "peer.example:65536", "0.0.0.0:444", "[fe80::1%eth0]:444", "224.0.0.1:444"} {
		if _, err := pinDirectPeer(t.Context(), input, lookup); err == nil {
			t.Fatal("invalid peer accepted", input)
		}
	}
	_, err := pinDirectPeer(t.Context(), "peer.example:444", func(context.Context, string, string) ([]netip.Addr, error) {
		return nil, errors.New("private DNS detail")
	})
	if err == nil || err.Error() != "tun: direct peer DNS resolution failed" {
		t.Fatal("unsafe DNS error", err)
	}
}
