// Фиктивное ядро только для lifecycle/recovery тестов; оно не реализует VPN.
package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"wireglow/internal/infrastructure/engines/manifest"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "-provider" || args[0] == "-d") {
		executable, err := os.Executable()
		if err != nil {
			os.Exit(5)
		}
		path := filepath.Join(filepath.Dir(executable), "engine-pids")
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			os.Exit(5)
		}
		_, err = f.WriteString(strconv.Itoa(os.Getpid()) + "\n")
		_ = f.Close()
		if err != nil {
			os.Exit(6)
		}
	}
	if len(args) == 1 && args[0] == "-h" {
		println("-provider -peer -listen -mode -transport -dns-mode -obf-profile -routes -debug -obf-timing -bond -links -manual-captcha -client-id")
		return
	}
	if len(args) > 0 && args[0] == "-provider" {
		for {
			time.Sleep(time.Hour)
		}
	}
	if len(args) > 0 && args[0] == "-t" {
		return
	}
	var path string
	for i, arg := range args {
		if arg == "-f" && i+1 < len(args) {
			path = args[i+1]
		}
	}
	var config struct {
		MixedPort  int    `json:"mixed-port"`
		Controller string `json:"external-controller"`
		Secret     string `json:"secret"`
	}
	data, err := os.ReadFile(path)
	if err != nil {
		os.Exit(2)
	}
	if json.Unmarshal(data, &config) != nil {
		os.Exit(3)
	}
	go func() {
		_ = http.ListenAndServe(config.Controller, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer "+config.Secret {
				w.WriteHeader(401)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"version": manifest.Version("mihomo")})
		}))
	}()
	server := &http.Server{Addr: "127.0.0.1:" + strconv.Itoa(config.MixedPort), Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "CONNECT" || !strings.HasPrefix(r.Host, "127.0.0.1:") {
			w.WriteHeader(403)
			return
		}
		destination, err := net.DialTimeout("tcp", r.Host, time.Second)
		if err != nil {
			w.WriteHeader(502)
			return
		}
		client, buffer, err := w.(http.Hijacker).Hijack()
		if err != nil {
			destination.Close()
			return
		}
		_, _ = buffer.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
		_ = buffer.Flush()
		go func() { _, _ = io.Copy(destination, client); destination.Close() }()
		_, _ = io.Copy(client, destination)
		client.Close()
	})}
	listener, err := net.Listen("tcp4", server.Addr)
	if err != nil {
		os.Exit(4)
	}
	if server.Serve(mixedListener{listener}) != nil {
		os.Exit(4)
	}
}

// Фикстура отвечает на UDP DNS probe локально, без обращения к внешнему resolver.
type bufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c bufferedConn) Read(p []byte) (int, error) { return c.reader.Read(p) }

type mixedListener struct{ net.Listener }

func (l mixedListener) Accept() (net.Conn, error) {
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		reader := bufio.NewReader(conn)
		prefix, err := reader.Peek(1)
		if err != nil {
			conn.Close()
			continue
		}
		_ = conn.SetReadDeadline(time.Time{})
		buffered := bufferedConn{conn, reader}
		if prefix[0] == 5 {
			go serveUDPProbe(buffered)
			continue
		}
		return buffered, nil
	}
}

func serveUDPProbe(control net.Conn) {
	defer control.Close()
	_ = control.SetDeadline(time.Now().Add(5 * time.Second))
	var greeting [3]byte
	if _, err := io.ReadFull(control, greeting[:]); err != nil || greeting != [3]byte{5, 1, 0} {
		return
	}
	if _, err := control.Write([]byte{5, 0}); err != nil {
		return
	}
	var request [10]byte
	if _, err := io.ReadFull(control, request[:]); err != nil || request != [10]byte{5, 3, 0, 1} {
		return
	}
	udp, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		return
	}
	defer udp.Close()
	_ = udp.SetDeadline(time.Now().Add(5 * time.Second))
	reply := []byte{5, 0, 0, 1, 127, 0, 0, 1, 0, 0}
	binary.BigEndian.PutUint16(reply[8:], uint16(udp.LocalAddr().(*net.UDPAddr).Port))
	if _, err := control.Write(reply); err != nil {
		return
	}
	buffer := make([]byte, 4096)
	n, peer, err := udp.ReadFromUDP(buffer)
	if err != nil || n < 22 || string(buffer[:10]) != string([]byte{0, 0, 0, 1, 1, 1, 1, 1, 0, 53}) {
		return
	}
	response := buffer[:n]
	response[12], response[13], response[17] = 0x81, 0x80, 1
	response = append(response, 0xc0, 0x0c, 0, 1, 0, 1, 0, 0, 0, 30, 0, 4, 192, 0, 2, 1)
	_, _ = udp.WriteToUDP(response, peer)
}
