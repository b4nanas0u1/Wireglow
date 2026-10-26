package tun

import (
	"errors"
	"net/netip"
	"strconv"
)

// Direct sessions have a fixed destination, unlike TURN. Match it rather than
// ephemeral source ports so even the first packet of a replacement socket bypasses TUN.
type directRoute struct {
	Peer string `json:"peer"`
	UID  uint32 `json:"uid"`
}

func (r directRoute) valid() bool {
	p, err := netip.ParseAddrPort(r.Peer)
	return err == nil && p.String() == r.Peer && p.Port() != 0 && r.UID != ^uint32(0) &&
		!p.Addr().Is4In6() && !p.Addr().IsUnspecified() && !p.Addr().IsMulticast() && !p.Addr().IsLinkLocalUnicast() && p.Addr().Zone() == ""
}
func (r directRoute) args(operation string) []string {
	p := netip.MustParseAddrPort(r.Peer)
	family := "-6"
	if p.Addr().Is4() {
		family = "-4"
	}
	uid := strconv.FormatUint(uint64(r.UID), 10)
	return []string{family, "rule", operation, "pref", "10216", "to", p.Addr().String(), "uidrange", uid + "-" + uid, "ipproto", "udp", "dport", strconv.Itoa(int(p.Port())), "lookup", "main"}
}
func directRoutes(peer string, uid uint32) ([]directRoute, error) {
	if peer == "" {
		return nil, nil
	}
	r := directRoute{Peer: peer, UID: uid}
	if !r.valid() {
		return nil, errors.New("tun: invalid direct peer exemption")
	}
	// Local destinations already use the kernel's higher-priority local table.
	if netip.MustParseAddrPort(peer).Addr().IsLoopback() {
		return nil, nil
	}
	return []directRoute{r}, nil
}
