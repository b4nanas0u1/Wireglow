package tun

import (
	"errors"
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/sys/windows"
	"net"
	"sort"
	"strings"
	"time"
	"unsafe"
)

type windowsLink struct {
	Name        string
	GUID        string
	Index       uint32
	Kind        uint32
	Up          bool
	Description string
	Gateways    []string
	Addresses   []string
	DNS         []string
}

func windowsLinks() ([]windowsLink, error) {
	var size uint32 = 16384
	for attempt := 0; attempt < 4; attempt++ {
		if size > 4<<20 {
			return nil, errors.New("tun: adapter inventory too large")
		}
		buffer := make([]byte, size)
		first := (*windows.IpAdapterAddresses)(unsafe.Pointer(&buffer[0]))
		err := windows.GetAdaptersAddresses(windows.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_GATEWAYS, 0, first, &size)
		if errors.Is(err, windows.ERROR_BUFFER_OVERFLOW) {
			continue
		}
		if err != nil {
			return nil, errors.New("tun: cannot inspect network adapters")
		}
		var links []windowsLink
		for a := first; a != nil; a = a.Next {
			l := windowsLink{Name: windows.UTF16PtrToString(a.FriendlyName), GUID: windows.BytePtrToString(a.AdapterName), Index: a.IfIndex, Kind: a.IfType, Up: a.OperStatus == windows.IfOperStatusUp, Description: windows.UTF16PtrToString(a.Description)}
			for g := a.FirstGatewayAddress; g != nil; g = g.Next {
				l.Gateways = append(l.Gateways, g.Address.IP().String())
			}
			for u := a.FirstUnicastAddress; u != nil; u = u.Next {
				l.Addresses = append(l.Addresses, fmt.Sprintf("%s/%d", u.Address.IP(), u.OnLinkPrefixLength))
			}
			for d := a.FirstDnsServerAddress; d != nil; d = d.Next {
				l.DNS = append(l.DNS, d.Address.IP().String())
			}
			sort.Strings(l.Gateways)
			sort.Strings(l.Addresses)
			links = append(links, l)
		}
		return links, nil
	}
	return nil, errors.New("tun: network adapters keep changing")
}

func (l windowsLink) tunnel() bool {
	d := strings.ToLower(l.Description + " " + l.Name)
	return l.Kind == 53 || l.Kind == 131 || strings.Contains(d, "wintun") || strings.Contains(d, "wireguard") || strings.Contains(d, "amnezia") || strings.Contains(d, "tap-windows") || strings.Contains(d, "openvpn")
}

func selectWindowsNetwork(links []windowsLink, owned string) (windowsLink, error) {
	var candidates []windowsLink
	for _, l := range links {
		if !l.Up || l.Name == owned {
			continue
		}
		if l.tunnel() {
			return windowsLink{}, fmt.Errorf("tun: another VPN/tunnel is active (%s)", l.Name)
		}
		if len(l.Gateways) > 0 && l.Kind != 24 {
			candidates = append(candidates, l)
		}
	}
	if len(candidates) != 1 {
		return windowsLink{}, errors.New("tun: a single unambiguous physical uplink is required")
	}
	return candidates[0], nil
}

func windowsNetwork(owned string) (windowsLink, error) {
	links, err := windowsLinks()
	if err != nil {
		return windowsLink{}, err
	}
	return selectWindowsNetwork(links, owned)
}

func sameWindowsNetwork(a, b windowsLink) bool {
	return a.GUID == b.GUID && a.Index == b.Index && strings.Join(a.Gateways, ",") == strings.Join(b.Gateways, ",") && strings.Join(a.Addresses, ",") == strings.Join(b.Addresses, ",")
}

func windowsDNSReady() error {
	conn, err := net.DialTimeout("udp4", "198.18.0.2:53", 500*time.Millisecond)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	query := dnsmessage.Message{Header: dnsmessage.Header{ID: 0x5747, RecursionDesired: true}, Questions: []dnsmessage.Question{{Name: dnsmessage.MustNewName("wireglow-readiness.invalid."), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}}}
	wire, err := query.Pack()
	if err != nil {
		return err
	}
	if _, err = conn.Write(wire); err != nil {
		return err
	}
	var packet [2048]byte
	n, err := conn.Read(packet[:])
	if err != nil {
		return err
	}
	var answer dnsmessage.Message
	if err = answer.Unpack(packet[:n]); err != nil || !answer.Response || answer.ID != query.ID || len(answer.Answers) == 0 {
		return errors.New("tun: userspace DNS is not ready")
	}
	return nil
}
