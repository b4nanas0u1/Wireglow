package tun

import (
	"net"
	"net/netip"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestKeepaliveSocketOwnershipAndDestinationGuards(t *testing.T) {
	for _, peer := range []string{"127.0.0.1", "::1", "::ffff:127.0.0.1", "0.0.0.0", "::", "224.0.0.1", "fe80::1"} {
		if keepalivePeer(netip.MustParseAddr(peer).Unmap(), 443) {
			t.Fatal("unsafe destination", peer)
		}
	}
	if keepalivePeer(netip.MustParseAddr(BootstrapDNS), 53) {
		t.Fatal("bootstrap DNS socket selected")
	}
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer socket.Close()
	raw, err := socket.SyscallConn()
	if err != nil {
		t.Fatal(err)
	}
	err = raw.Control(func(fd uintptr) {
		var stat unix.Stat_t
		if err := unix.Fstat(int(fd), &stat); err != nil {
			t.Error(err)
			return
		}
		if ok, err := keepaliveSocket(int(fd), stat.Ino+1, uint32(os.Getuid())); ok || err != nil {
			t.Error("reused descriptor accepted", ok, err)
		}
		if ok, err := keepaliveSocket(int(fd), stat.Ino, uint32(os.Getuid()+1)); ok || err == nil {
			t.Error("wrong owner accepted", ok, err)
		}
		if ok, err := keepaliveSocket(int(fd), stat.Ino, uint32(os.Getuid())); ok || err != nil {
			t.Error("unconnected local listener accepted", ok, err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
