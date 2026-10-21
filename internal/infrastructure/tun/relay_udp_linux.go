package tun

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// UDP must retain its original source port. Sending a live relay socket through
// a second userspace proxy changes its peer-visible endpoint. UID plus the
// kernel-allocated port selects the socket; overlapping non-relay sockets are
// rejected before installing an exemption.
type relayUDPRoute struct {
	Family string `json:"family"`
	Port   uint16 `json:"port"`
	UID    uint32 `json:"uid"`
}

func (r relayUDPRoute) valid() bool {
	return (r.Family == "-4" || r.Family == "-6") && r.Port != 0 && r.UID != ^uint32(0)
}
func (r relayUDPRoute) args(operation string) []string {
	uid := strconv.FormatUint(uint64(r.UID), 10)
	return []string{r.Family, "rule", operation, "pref", "10217", "uidrange", uid + "-" + uid, "ipproto", "udp", "sport", strconv.Itoa(int(r.Port)), "lookup", "main"}
}
func relayUDPRoutes(pid int, ownerUID uint32) ([]relayUDPRoute, error) {
	base := filepath.Join("/proc", strconv.Itoa(pid))
	fds, err := os.ReadDir(filepath.Join(base, "fd"))
	if err != nil {
		return nil, errors.New("tun: cannot inspect relay UDP sockets")
	}
	owned := make(map[string]bool)
	for _, fd := range fds {
		target, err := os.Readlink(filepath.Join(base, "fd", fd.Name()))
		if err == nil && strings.HasPrefix(target, "socket:[") {
			owned[strings.TrimSuffix(strings.TrimPrefix(target, "socket:["), "]")] = true
		}
	}
	type socket struct {
		route    relayUDPRoute
		inode    string
		loopback bool
		dual     bool
	}
	var sockets []socket
	for _, table := range []string{"udp", "udp6"} {
		file, err := os.Open(filepath.Join(base, "net", table))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, errors.New("tun: cannot read relay UDP table")
		}
		scan := bufio.NewScanner(file)
		for scan.Scan() {
			fields := strings.Fields(scan.Text())
			if len(fields) < 10 || fields[0] == "sl" || fields[9] == "0" {
				continue
			}
			local, port, first := procAddress(fields[1])
			remote, _, second := procAddress(fields[2])
			uid, third := strconv.ParseUint(fields[7], 10, 32)
			if first != nil || second != nil || third != nil {
				file.Close()
				return nil, errors.New("tun: invalid UDP table")
			}
			family := "-6"
			if local.Is4() {
				family = "-4"
			}
			sockets = append(sockets, socket{relayUDPRoute{family, port, uint32(uid)}, fields[9], local.IsLoopback() || remote.IsLoopback(), table == "udp6" && local.IsUnspecified()})
		}
		err = scan.Err()
		file.Close()
		if err != nil {
			return nil, err
		}
	}
	seen := make(map[relayUDPRoute]bool)
	var routes []relayUDPRoute
	for _, s := range sockets {
		if !owned[s.inode] || s.loopback || !s.route.valid() {
			continue
		}
		if s.route.UID != ownerUID {
			return nil, errors.New("tun: relay UDP socket owner mismatch")
		}
		candidates := []relayUDPRoute{s.route}
		if s.dual {
			other := s.route
			other.Family = "-4"
			candidates = append(candidates, other)
		}
		for _, candidate := range candidates {
			for _, other := range sockets {
				if other.route.UID == candidate.UID && other.route.Port == candidate.Port && !other.loopback && !owned[other.inode] && (other.route.Family == candidate.Family || other.dual) {
					return nil, errors.New("tun: relay UDP port overlaps another process")
				}
			}
			if !seen[candidate] {
				if len(routes) >= 256 {
					return nil, errors.New("tun: too many relay UDP sockets")
				}
				seen[candidate] = true
				routes = append(routes, candidate)
			}
		}
	}
	return routes, nil
}
