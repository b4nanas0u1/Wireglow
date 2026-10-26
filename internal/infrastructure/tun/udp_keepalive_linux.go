package tun

import (
	"context"
	"crypto/rand"
	"errors"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// TURN Binding indications keep mihomo's 60-second UDP mapping alive. They are
// sent through the relay's own connected socket: using a separate socket would
// refresh a different mapping. pidfd prevents PID reuse; every borrowed fd is
// checked against its observed inode and UID and closed immediately after use.
// No socket options, flags, routes or firewall settings are changed here.
type udpKeepalive struct {
	pid, pidfd int
	uid        uint32
	cancel     context.CancelFunc
	done       chan struct{}
	err        error // published by closing done
}

func newUDPKeepalive(pid int, uid uint32) (*udpKeepalive, error) {
	fd, err := unix.PidfdOpen(pid, 0)
	if err != nil {
		return nil, errors.New("tun: TURN UDP keepalive requires pidfd support")
	}
	k := &udpKeepalive{pid: pid, pidfd: fd, uid: uid, done: make(chan struct{})}
	if err := k.send(); err != nil {
		unix.Close(fd)
		return nil, err
	}
	return k, nil
}
func (k *udpKeepalive) start() {
	ctx, cancel := context.WithCancel(context.Background())
	k.cancel = cancel
	go func() {
		defer close(k.done)
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			if err := k.send(); err != nil {
				k.err = err
				return
			}
		}
	}()
}
func (k *udpKeepalive) failure() error {
	select {
	case <-k.done:
		return k.err
	default:
		return nil
	}
}
func (k *udpKeepalive) close() {
	if k.cancel != nil {
		k.cancel()
		<-k.done
	}
	unix.Close(k.pidfd)
}
func (k *udpKeepalive) send() error {
	base := filepath.Join("/proc", strconv.Itoa(k.pid), "fd")
	entries, err := os.ReadDir(base)
	if err != nil {
		return errors.New("tun: TURN UDP relay inspection failed")
	}
	var payload [20]byte
	if _, err := rand.Read(payload[8:]); err != nil {
		return errors.New("tun: TURN keepalive randomness failed")
	}
	// RFC 8489 Binding indication, empty body, no response requested.
	copy(payload[:8], []byte{0, 0x11, 0, 0, 0x21, 0x12, 0xa4, 0x42})
	count := 0
	for _, entry := range entries {
		target, err := os.Readlink(filepath.Join(base, entry.Name()))
		if err != nil || !strings.HasPrefix(target, "socket:[") || !strings.HasSuffix(target, "]") {
			continue
		}
		inode, err := strconv.ParseUint(strings.TrimSuffix(strings.TrimPrefix(target, "socket:["), "]"), 10, 64)
		if err != nil {
			continue
		}
		number, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		fd, err := unix.PidfdGetfd(k.pidfd, number, 0)
		if errors.Is(err, unix.EBADF) {
			// The socket closed during enumeration.
			continue
		}
		if errors.Is(err, unix.ESRCH) {
			return errors.New("tun: TURN UDP relay exited")
		}
		if err != nil {
			return errors.New("tun: TURN UDP keepalive socket access denied")
		}
		eligible, err := keepaliveSocket(fd, inode, k.uid)
		if err == nil && eligible {
			count++
			if count > 256 {
				unix.Close(fd)
				return errors.New("tun: too many TURN UDP sockets")
			}
			// Do not change the shared file description's O_NONBLOCK or deadlines.
			sendErr := unix.Sendto(fd, payload[:], unix.MSG_DONTWAIT|unix.MSG_NOSIGNAL, nil)
			// A dead TURN candidate may return an ICMP error while other streams
			// are healthy. Let FreeTurn handle transport failures; another tick
			// retries transient send errors. Permission loss cannot be ignored.
			if errors.Is(sendErr, unix.EACCES) || errors.Is(sendErr, unix.EPERM) {
				unix.Close(fd)
				return errors.New("tun: TURN UDP keepalive send denied")
			}
		}
		unix.Close(fd)
		if err != nil {
			return err
		}
	}
	return nil
}
func keepaliveSocket(fd int, inode uint64, uid uint32) (bool, error) {
	var stat unix.Stat_t
	if unix.Fstat(fd, &stat) != nil || stat.Ino != inode {
		return false, nil
	}
	if stat.Uid != uid {
		return false, errors.New("tun: TURN UDP socket owner changed")
	}
	typeID, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_TYPE)
	if err != nil || typeID != unix.SOCK_DGRAM {
		return false, nil
	}
	protocol, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_PROTOCOL)
	if err != nil || protocol != unix.IPPROTO_UDP {
		return false, nil
	}
	peer, err := unix.Getpeername(fd)
	if err != nil {
		// Local listener or disconnected socket.
		return false, nil
	}
	switch peer := peer.(type) {
	case *unix.SockaddrInet4:
		return keepalivePeer(netip.AddrFrom4(peer.Addr), peer.Port), nil
	case *unix.SockaddrInet6:
		return peer.ZoneId == 0 && keepalivePeer(netip.AddrFrom16(peer.Addr).Unmap(), peer.Port), nil
	}
	return false, nil
}

func keepalivePeer(address netip.Addr, port int) bool {
	return port > 0 && !address.IsLoopback() && !address.IsUnspecified() && !address.IsMulticast() && !address.IsLinkLocalUnicast() && !(port == 53 && address.String() == BootstrapDNS)
}
