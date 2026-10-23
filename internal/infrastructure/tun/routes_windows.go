package tun

import (
	"encoding/binary"
	"errors"
	"golang.org/x/sys/windows"
	"net/netip"
	"sort"
	"unsafe"
)

var ipHelper = windows.NewLazySystemDLL("iphlpapi.dll")

// Layouts of SOCKADDR_INET, IP_ADDRESS_PREFIX and MIB_IPFORWARD_ROW2 from
// Windows SDK netioapi.h. No route mutation uses these read-only bindings.
type routeAddress struct {
	Family uint16
	Data   [26]byte
}
type routePrefix struct {
	Address routeAddress
	Length  uint8
	_       [3]byte
}
type windowsRoute struct {
	LUID                               uint64
	Index                              uint32
	Destination                        routePrefix
	NextHop                            routeAddress
	SitePrefix                         uint8
	Valid, Preferred, Metric, Protocol uint32
	Loopback, Auto, Publish, Immortal  uint8
	Age, Origin                        uint32
}
type routeTable struct {
	Count uint32
	Rows  [1]windowsRoute
}

func windowsRoutes() ([]windowsRoute, error) {
	var table *routeTable
	code, _, _ := ipHelper.NewProc("GetIpForwardTable2").Call(windows.AF_UNSPEC, uintptr(unsafe.Pointer(&table)))
	if code != 0 {
		return nil, errors.New("tun: cannot inspect routing table")
	}
	defer ipHelper.NewProc("FreeMibTable").Call(uintptr(unsafe.Pointer(table)))
	if table == nil || table.Count > 100000 {
		return nil, errors.New("tun: invalid routing table")
	}
	return append([]windowsRoute(nil), unsafe.Slice(&table.Rows[0], int(table.Count))...), nil
}

func (a routeAddress) ip() netip.Addr {
	if a.Family == windows.AF_INET {
		var b [4]byte
		copy(b[:], a.Data[2:6])
		return netip.AddrFrom4(b)
	}
	if a.Family == windows.AF_INET6 {
		var b [16]byte
		copy(b[:], a.Data[6:22])
		return netip.AddrFrom16(b)
	}
	return netip.Addr{}
}

func windowsCaptureReady(index uint32) error {
	routes, err := windowsRoutes()
	if err != nil {
		return err
	}
	v4, v6 := false, false
	for _, r := range routes {
		if r.Index != index {
			continue
		}
		// BuildAutoRouteRanges splits exclusions into CIDRs, so a single /0
		// need not exist. Both families must have global capture coverage.
		p := netip.PrefixFrom(r.Destination.Address.ip(), int(r.Destination.Length))
		if p.Contains(netip.MustParseAddr("203.0.113.1")) {
			v4 = true
		}
		if p.Contains(netip.MustParseAddr("2001:db8::1")) {
			v6 = true
		}
	}
	if !v4 || !v6 {
		return errors.New("tun: IPv4/IPv6 capture routes are not ready")
	}
	return nil
}

func windowsTCPPeers(pid int) ([]string, error) {
	var result []string
	for _, family := range []uint32{windows.AF_INET, windows.AF_INET6} {
		var size uint32
		proc := ipHelper.NewProc("GetExtendedTcpTable")
		code, _, _ := proc.Call(0, uintptr(unsafe.Pointer(&size)), 0, uintptr(family), 5, 0) // TCP_TABLE_OWNER_PID_ALL
		if code != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) && code != 0 {
			return nil, errors.New("tun: cannot inspect relay TCP sockets")
		}
		if size < 4 || size > 16<<20 {
			return nil, errors.New("tun: invalid TCP table size")
		}
		buffer := make([]byte, size)
		code, _, _ = proc.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&size)), 0, uintptr(family), 5, 0)
		if code != 0 {
			return nil, errors.New("tun: TCP sockets changed during inspection")
		}
		count := int(binary.LittleEndian.Uint32(buffer[:4]))
		rowSize := 24
		if family == windows.AF_INET6 {
			rowSize = 56
		}
		if count > (len(buffer)-4)/rowSize {
			return nil, errors.New("tun: malformed TCP table")
		}
		for i := 0; i < count; i++ {
			row := buffer[4+i*rowSize : 4+(i+1)*rowSize]
			if int(binary.LittleEndian.Uint32(row[rowSize-4:])) != pid {
				continue
			}
			var ip netip.Addr
			if family == windows.AF_INET {
				var a [4]byte
				copy(a[:], row[12:16])
				ip = netip.AddrFrom4(a)
			} else {
				var a [16]byte
				copy(a[:], row[24:40])
				ip = netip.AddrFrom16(a)
			}
			if ip.IsGlobalUnicast() {
				result = append(result, ip.String())
			}
			if len(result) > 256 {
				return nil, errors.New("tun: too many relay TCP destinations")
			}
		}
	}
	sort.Strings(result)
	return result, nil
}
