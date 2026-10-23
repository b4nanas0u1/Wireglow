//go:build linux || windows

package tun

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"sync/atomic"
	"time"
)

func fakeSOCKS(listener net.Listener, tcpCount, udpCount *atomic.Int32, observers ...func(byte, []byte, uint16)) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go func() {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(10 * time.Second))
			var h [4]byte
			if _, err := io.ReadFull(conn, h[:2]); err != nil {
				return
			}
			methods := make([]byte, int(h[1]))
			if _, err := io.ReadFull(conn, methods); err != nil {
				return
			}
			conn.Write([]byte{5, 0})
			if _, err := io.ReadFull(conn, h[:]); err != nil {
				return
			}
			command := h[1]
			var address []byte
			switch h[3] {
			case 1:
				address = make([]byte, 4)
			case 4:
				address = make([]byte, 16)
			case 3:
				var l [1]byte
				io.ReadFull(conn, l[:])
				address = make([]byte, int(l[0]))
			default:
				return
			}
			if _, err := io.ReadFull(conn, address); err != nil {
				return
			}
			var port [2]byte
			if _, err := io.ReadFull(conn, port[:]); err != nil {
				return
			}
			if command == 1 {
				for _, observe := range observers {
					observe(1, address, binary.BigEndian.Uint16(port[:]))
				}
				tcpCount.Add(1)
				conn.Write([]byte{5, 0, 0, 1, 127, 0, 0, 1, 0, 0})
				reader := bufio.NewReader(conn)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					if line == "\r\n" {
						break
					}
				}
				io.WriteString(conn, "HTTP/1.1 200 OK\r\nContent-Length: 6\r\nConnection: close\r\n\r\ntun-ok")
				return
			}
			if command != 3 {
				return
			}
			socket, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
			if err != nil {
				return
			}
			defer socket.Close()
			socket.SetDeadline(time.Now().Add(8 * time.Second))
			reply := []byte{5, 0, 0, 1, 127, 0, 0, 1, 0, 0}
			binary.BigEndian.PutUint16(reply[8:], uint16(socket.LocalAddr().(*net.UDPAddr).Port))
			conn.Write(reply)
			var packet [2048]byte
			for {
				n, peer, err := socket.ReadFromUDP(packet[:])
				if err != nil {
					return
				}
				udpCount.Add(1)
				if n >= 10 && packet[3] == 1 {
					for _, observe := range observers {
						observe(3, packet[4:8], binary.BigEndian.Uint16(packet[8:10]))
					}
				}
				if n >= 22 && packet[3] == 4 {
					for _, observe := range observers {
						observe(3, packet[4:20], binary.BigEndian.Uint16(packet[20:22]))
					}
				}
				_, _ = socket.WriteToUDP(packet[:n], peer)
			}
		}()
	}
}
