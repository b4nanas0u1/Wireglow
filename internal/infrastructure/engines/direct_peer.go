package engines

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"time"
)

// Resolve once before starting FreeTurn: both its transport and the helper must
// use the same destination even if the profile's DNS answer later changes.
func pinDirectPeer(ctx context.Context, peer string, lookup func(context.Context, string, string) ([]netip.Addr, error)) (string, error) {
	host, port, err := net.SplitHostPort(peer)
	number, portErr := strconv.ParseUint(port, 10, 16)
	if err != nil || host == "" || portErr != nil || number == 0 {
		return "", errors.New("tun: invalid direct peer")
	}
	address, err := netip.ParseAddr(host)
	if err != nil {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		addresses, err := lookup(ctx, "ip", host)
		if err != nil || len(addresses) == 0 {
			return "", errors.New("tun: direct peer DNS resolution failed")
		}
		address = addresses[0]
		for _, candidate := range addresses {
			if candidate.Is4() || candidate.Is4In6() {
				address = candidate
				break
			}
		}
	}
	address = address.Unmap()
	if !address.IsValid() || address.IsUnspecified() || address.IsMulticast() || address.IsLinkLocalUnicast() || address.Zone() != "" {
		return "", errors.New("tun: invalid direct peer IP")
	}
	return netip.AddrPortFrom(address, uint16(number)).String(), nil
}
