package engines

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/netip"
	"strings"
	"time"
)

// ProbeUDP подтверждает DNS-ответ через SOCKS5 UDP ASSOCIATE и тот же PROXY outbound.
func ProbeUDP(ctx context.Context, mixed string) error {
	if _, err := loopbackPort(mixed); err != nil {
		return err
	}
	failure := probeError("probe: UDP SOCKS association failed")
	parentCtx := ctx
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	control, err := (&net.Dialer{}).DialContext(ctx, "tcp4", mixed)
	if err != nil {
		return failure
	}
	defer control.Close()
	stop := context.AfterFunc(ctx, func() { _ = control.Close() })
	defer stop()
	deadline, _ := ctx.Deadline()
	_ = control.SetDeadline(deadline)
	if _, err := control.Write([]byte{5, 1, 0}); err != nil {
		return failure
	}
	var hello [2]byte
	if _, err := io.ReadFull(control, hello[:]); err != nil || hello != [2]byte{5, 0} {
		return failure
	}
	if _, err := control.Write([]byte{5, 3, 0, 1, 0, 0, 0, 0, 0, 0}); err != nil {
		return failure
	}
	var reply [4]byte
	if _, err := io.ReadFull(control, reply[:]); err != nil || reply[0] != 5 || reply[1] != 0 || reply[2] != 0 {
		return failure
	}
	var relay netip.Addr
	switch reply[3] {
	case 1:
		var addr [4]byte
		if _, err := io.ReadFull(control, addr[:]); err != nil {
			return failure
		}
		relay = netip.AddrFrom4(addr)
	case 4:
		var addr [16]byte
		if _, err := io.ReadFull(control, addr[:]); err != nil {
			return failure
		}
		relay = netip.AddrFrom16(addr)
	default:
		return failure
	}
	var port [2]byte
	if _, err := io.ReadFull(control, port[:]); err != nil {
		return failure
	}
	p := binary.BigEndian.Uint16(port[:])
	if p == 0 {
		return failure
	}
	if relay.IsUnspecified() {
		relay = netip.MustParseAddr("127.0.0.1")
	}
	if !relay.IsLoopback() {
		return failure
	}
	packetConn, err := (&net.Dialer{}).DialContext(ctx, "udp", netip.AddrPortFrom(relay, p).String())
	if err != nil {
		return failure
	}
	defer packetConn.Close()
	stopUDP := context.AfterFunc(ctx, func() { _ = packetConn.Close() })
	defer stopUDP()
	_ = packetConn.SetDeadline(deadline)
	query := make([]byte, 12)
	if _, err := rand.Read(query[:2]); err != nil {
		return err
	}
	query[2] = 1
	query[5] = 1
	for _, label := range strings.Split("example.org", ".") {
		query = append(query, byte(len(label)))
		query = append(query, label...)
	}
	query = append(query, 0, 0, 1, 0, 1)
	// Один фиксированный resolver через PROXY, без fallback в физическую сеть.
	failure = probeError("probe: UDP DNS request failed")
	header := []byte{0, 0, 0, 1, 1, 1, 1, 1, 0, 53}
	if _, err := packetConn.Write(append(header, query...)); err != nil {
		return failure
	}
	buffer := make([]byte, 4096)
	n, err := packetConn.Read(buffer)
	if err != nil {
		if parentCtx.Err() != nil {
			return probeError("probe: UDP DNS cancelled or startup deadline exceeded")
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return probeError("probe: UDP DNS timeout after 10s")
		}
		var timeout net.Error
		if errors.As(err, &timeout) && timeout.Timeout() {
			return probeError("probe: UDP DNS timeout after 10s")
		}
		return failure
	}
	failure = probeError("probe: UDP DNS response invalid")
	if n < 22 {
		return failure
	}
	if string(buffer[:10]) != string(header) {
		return failure
	}
	dns := buffer[10:n]
	if !validDNSAnswer(dns, query) {
		return failure
	}
	return nil
}

func validDNSAnswer(response, query []byte) bool {
	if len(response) < len(query) || len(query) < 12 || string(response[:2]) != string(query[:2]) || response[2]&0xf8 != 0x80 || response[2]&2 != 0 || response[3]&15 != 0 || binary.BigEndian.Uint16(response[4:6]) != 1 || string(response[12:len(query)]) != string(query[12:]) {
		return false
	}
	count := int(binary.BigEndian.Uint16(response[6:8]))
	if count == 0 || count > 256 {
		return false
	}
	position := len(query)
	found := false
	for range count {
		for {
			if position >= len(response) {
				return false
			}
			size := int(response[position])
			position++
			if size == 0 {
				break
			}
			if size&0xc0 == 0xc0 {
				if position >= len(response) {
					return false
				}
				offset := (size&0x3f)<<8 | int(response[position])
				position++
				if offset >= len(response) {
					return false
				}
				break
			}
			if size > 63 || position+size > len(response) {
				return false
			}
			position += size
		}
		if position+10 > len(response) {
			return false
		}
		kind := binary.BigEndian.Uint16(response[position:])
		class := binary.BigEndian.Uint16(response[position+2:])
		size := int(binary.BigEndian.Uint16(response[position+8:]))
		position += 10
		if position+size > len(response) {
			return false
		}
		if kind == 1 && class == 1 && size == 4 {
			found = true
		}
		position += size
	}
	return found
}
