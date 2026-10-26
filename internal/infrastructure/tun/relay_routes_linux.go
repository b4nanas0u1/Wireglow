package tun

import (
	"bufio"
	"encoding/hex"
	"errors"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Existing TCP streams must bypass the userspace TUN before reaching its TCP
// stack: PROCESS-PATH cannot rescue an ACK for a stream whose SYN it never saw.
type relayRoute struct {
	Family     string `json:"family"`
	Local      string `json:"local"`
	Remote     string `json:"remote"`
	LocalPort  uint16 `json:"localPort"`
	RemotePort uint16 `json:"remotePort"`
}

func (r relayRoute) valid() bool {
	local, e1 := netip.ParseAddr(r.Local)
	remote, e2 := netip.ParseAddr(r.Remote)
	if e1 != nil || e2 != nil || local.IsLoopback() || remote.IsLoopback() || local.IsUnspecified() || remote.IsUnspecified() || r.LocalPort == 0 || r.RemotePort == 0 {
		return false
	}
	return (r.Family == "-4" && local.Is4() && remote.Is4()) || (r.Family == "-6" && local.Is6() && remote.Is6())
}
func (r relayRoute) args(operation string) []string {
	return []string{r.Family, "rule", operation, "pref", "10218", "from", r.Local, "to", r.Remote, "ipproto", "tcp", "sport", strconv.Itoa(int(r.LocalPort)), "dport", strconv.Itoa(int(r.RemotePort)), "lookup", "main"}
}
func procAddress(text string) (netip.Addr, uint16, error) {
	address, port, ok := strings.Cut(text, ":")
	if !ok {
		return netip.Addr{}, 0, errors.New("tun: invalid socket address")
	}
	data, err := hex.DecodeString(address)
	if err != nil || len(data) != 4 && len(data) != 16 {
		return netip.Addr{}, 0, errors.New("tun: invalid socket IP")
	}
	// Supported Linux target is x64; /proc encodes each native 32-bit word.
	for i := 0; i < len(data); i += 4 {
		data[i], data[i+3] = data[i+3], data[i]
		data[i+1], data[i+2] = data[i+2], data[i+1]
	}
	ip, ok := netip.AddrFromSlice(data)
	if !ok {
		return netip.Addr{}, 0, errors.New("tun: invalid socket IP")
	}
	number, err := strconv.ParseUint(port, 16, 16)
	return ip.Unmap(), uint16(number), err
}
func relayTCPRoutes(pid int) ([]relayRoute, error) {
	base := filepath.Join("/proc", strconv.Itoa(pid))
	files, err := os.ReadDir(filepath.Join(base, "fd"))
	if err != nil {
		return nil, errors.New("tun: cannot inspect relay sockets")
	}
	inodes := make(map[string]bool)
	for _, file := range files {
		target, err := os.Readlink(filepath.Join(base, "fd", file.Name()))
		if err != nil {
			continue
		}
		if strings.HasPrefix(target, "socket:[") && strings.HasSuffix(target, "]") {
			inodes[strings.TrimSuffix(strings.TrimPrefix(target, "socket:["), "]")] = true
		}
	}
	var routes []relayRoute
	seen := make(map[relayRoute]bool)
	for _, table := range []string{"tcp", "tcp6"} {
		f, err := os.Open(filepath.Join(base, "net", table))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, errors.New("tun: cannot inspect TCP table")
		}
		scan := bufio.NewScanner(f)
		for scan.Scan() {
			fields := strings.Fields(scan.Text())
			if len(fields) < 10 || !inodes[fields[9]] || fields[3] == "0A" {
				continue
			}
			local, lp, e1 := procAddress(fields[1])
			remote, rp, e2 := procAddress(fields[2])
			if e1 != nil || e2 != nil {
				f.Close()
				return nil, errors.New("tun: invalid TCP table")
			}
			family := "-6"
			if local.Is4() {
				family = "-4"
			}
			r := relayRoute{family, local.String(), remote.String(), lp, rp}
			if r.valid() && !seen[r] {
				if len(routes) >= 256 {
					f.Close()
					return nil, errors.New("tun: too many relay TCP sockets")
				}
				routes = append(routes, r)
				seen[r] = true
			}
		}
		err = scan.Err()
		f.Close()
		if err != nil {
			return nil, err
		}
	}
	return routes, nil
}
