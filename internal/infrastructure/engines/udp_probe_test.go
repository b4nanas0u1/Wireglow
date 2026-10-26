package engines

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

func TestUDPProbeChecksRealSOCKSAssociationAndDNSAnswer(t *testing.T) {
	for _, valid := range []bool{true, false} {
		t.Run(map[bool]string{true: "valid", false: "wrong transaction"}[valid], func(t *testing.T) {
			listener, err := net.Listen("tcp4", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			udp, err := net.ListenPacket("udp4", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer udp.Close()
			_, portText, _ := net.SplitHostPort(udp.LocalAddr().String())
			port, err := loopbackPort("127.0.0.1:" + portText)
			if err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() {
				control, err := listener.Accept()
				if err != nil {
					done <- err
					return
				}
				defer control.Close()
				_ = control.SetDeadline(time.Now().Add(2 * time.Second))
				var greeting [3]byte
				if _, err := io.ReadFull(control, greeting[:]); err != nil {
					done <- err
					return
				}
				if greeting != [3]byte{5, 1, 0} {
					done <- io.ErrUnexpectedEOF
					return
				}
				_, _ = control.Write([]byte{5, 0})
				var request [10]byte
				if _, err := io.ReadFull(control, request[:]); err != nil {
					done <- err
					return
				}
				if request[1] != 3 {
					done <- io.ErrUnexpectedEOF
					return
				}
				reply := []byte{5, 0, 0, 1, 127, 0, 0, 1, 0, 0}
				binary.BigEndian.PutUint16(reply[8:], uint16(port))
				_, _ = control.Write(reply)
				buffer := make([]byte, 4096)
				_ = udp.SetDeadline(time.Now().Add(2 * time.Second))
				n, peer, err := udp.ReadFrom(buffer)
				if err != nil {
					done <- err
					return
				}
				response := append([]byte(nil), buffer[:n]...)
				response[12] = 0x81
				response[13] = 0x80
				response[17] = 1
				response = append(response, 0xc0, 0x0c, 0, 1, 0, 1, 0, 0, 0, 30, 0, 4, 192, 0, 2, 1)
				if !valid {
					response[10] ^= 1
				}
				_, err = udp.WriteTo(response, peer)
				done <- err
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			err = ProbeUDP(ctx, listener.Addr().String())
			if (err == nil) != valid {
				t.Fatal("incorrect probe result", err)
			}
			if err := <-done; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDNSProbeRejectsTruncatedAndEmptyAnswers(t *testing.T) {
	query := []byte{1, 2, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 'x', 0, 0, 1, 0, 1}
	response := append([]byte(nil), query...)
	response[2] = 0x81
	response[3] = 0x80
	response[7] = 1
	if validDNSAnswer(response, query) {
		t.Fatal("missing answer accepted")
	}
	response = append(response, 0xc0, 12, 0, 1, 0, 1, 0, 0, 0, 30, 0, 4, 192, 0, 2, 1)
	if !validDNSAnswer(response, query) {
		t.Fatal("valid DNS answer rejected")
	}
	response[2] |= 2
	if validDNSAnswer(response, query) {
		t.Fatal("truncated answer accepted")
	}
}
