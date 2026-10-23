package engines

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"wireglow/internal/application/ports"
	"wireglow/internal/domain/profile"
	"wireglow/internal/domain/proxymode"
	"wireglow/internal/freeturn"
	"wireglow/internal/infrastructure/engines/manifest"
	"wireglow/internal/infrastructure/persistence"
	"wireglow/internal/infrastructure/tun"
)

type Binary struct {
	Path   string
	SHA256 string
}
type Preparer struct {
	Relay       Binary
	Mihomo      Binary
	RuntimeDir  string
	ProbeURL    string
	ProbeStatus int
	Ports       Ports
	Addresses   func() (relay, mixed, api string, err error)
}

// Ports задаёт локальные endpoints ядер. Нулевые значения используют production defaults.
type Ports struct {
	Relay int
	Mixed int
	API   int
}

const DefaultRelayAddress = "127.0.0.1:17890"
const DefaultProxyAddress = "127.0.0.1:17891"
const DefaultAPIPort = 17892

var defaultPorts = Ports{Relay: 17890, Mixed: 17891, API: DefaultAPIPort}

func ValidateAddresses(relay, mixed string, apiPort int) error {
	relayPort, err := loopbackPort(relay)
	if err != nil {
		return err
	}
	mixedPort, err := loopbackPort(mixed)
	if err != nil {
		return err
	}
	if apiPort < 1 || apiPort > 65535 {
		return errors.New("engine: invalid API port")
	}
	if relayPort == mixedPort || relayPort == apiPort || mixedPort == apiPort {
		return errors.New("engine: local ports must differ")
	}
	return nil
}

func (p Ports) addresses() (relay, mixed, api string, err error) {
	if p == (Ports{}) {
		p = defaultPorts
	}
	if p.Relay < 1 || p.Relay > 65535 || p.Mixed < 1 || p.Mixed > 65535 || p.API < 1 || p.API > 65535 || p.Relay == p.Mixed || p.Relay == p.API || p.Mixed == p.API {
		return "", "", "", errors.New("engine: invalid or colliding ports")
	}
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(p.Relay)), net.JoinHostPort("127.0.0.1", strconv.Itoa(p.Mixed)), net.JoinHostPort("127.0.0.1", strconv.Itoa(p.API)), nil
}

type dnsConfig struct {
	Enable      bool     `json:"enable"`
	IPv6        bool     `json:"ipv6"`
	Mode        string   `json:"enhanced-mode"`
	Nameservers []string `json:"nameserver"`
}

func tunDNS() *dnsConfig {
	// The outer TUN owns fake-IP mappings. Inner WG must never resolve through
	// the captured system resolver and receive addresses from that other cache.
	return &dnsConfig{Enable: true, Mode: "redir-host", Nameservers: []string{"https://1.0.0.1/dns-query#PROXY"}}
}

type config struct {
	DNS         *dnsConfig `json:"dns,omitempty"`
	MixedPort   int        `json:"mixed-port"`
	BindAddress string     `json:"bind-address"`
	AllowLAN    bool       `json:"allow-lan"`
	Mode        string     `json:"mode"`
	IPv6        bool       `json:"ipv6"`
	LogLevel    string     `json:"log-level"`
	Controller  string     `json:"external-controller"`
	Secret      string     `json:"secret"`
	Proxies     []Outbound `json:"proxies"`
	Rules       []string   `json:"rules"`
	Tun         struct {
		Enable bool `json:"enable"`
	} `json:"tun"`
}

func checkBinary(binary Binary) error {
	if !filepath.IsAbs(binary.Path) {
		return errors.New("engine: absolute executable path required")
	}
	f, err := os.Open(binary.Path)
	if err != nil {
		return errors.New("engine: executable unavailable")
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil || !stat.Mode().IsRegular() {
		return errors.New("engine: executable is not a regular file")
	}
	if binary.SHA256 != "" {
		expected, err := hex.DecodeString(binary.SHA256)
		if err != nil || len(expected) != 32 {
			return errors.New("engine: invalid binary digest")
		}
		hash := sha256.New()
		if _, err := io.Copy(hash, f); err != nil {
			return err
		}
		if hex.EncodeToString(hash.Sum(nil)) != strings.ToLower(binary.SHA256) {
			return errors.New("engine: binary digest mismatch")
		}
	}
	return nil
}

// freePorts держит резервные sockets одновременно, чтобы endpoints не совпали.
func freePorts() (relay, mixed, api string, release func(), err error) {
	var listeners []net.Listener
	release = func() {
		for _, l := range listeners {
			_ = l.Close()
		}
	}
	for range 3 {
		l, e := net.Listen("tcp4", "127.0.0.1:0")
		if e != nil {
			release()
			return "", "", "", nil, e
		}
		listeners = append(listeners, l)
	}
	return listeners[0].Addr().String(), listeners[1].Addr().String(), listeners[2].Addr().String(), release, nil
}

func (p *Preparer) Prepare(ctx context.Context, profile profile.Profile) (*ports.Plan, error) {
	return p.PrepareForMode(ctx, profile, proxymode.Local)
}

func (p *Preparer) PrepareForMode(ctx context.Context, profile profile.Profile, mode proxymode.Mode) (_ *ports.Plan, err error) {
	directPeer := ""
	if mode == proxymode.Tun {
		if runtime.GOOS == "windows" && profile.Connection.Provider == freeturn.ProviderVK && profile.Connection.TurnTransport == freeturn.TurnTransportUDP {
			return nil, windowsTURNUDPError{}
		}
		// Override only the runtime copy. Stored/imported profile remains intact.
		profile.Connection.DNS = freeturn.DNSSettings{Mode: freeturn.DNSModePlain, Servers: []string{tun.BootstrapDNS}}
		if runtime.GOOS == "windows" {
			// Windows strict-route rejects physical port 53. The pinned relay's
			// DoH resolver has numeric bootstrap addresses and no plain fallback.
			profile.Connection.DNS = freeturn.DNSSettings{Mode: freeturn.DNSModeDoH}
		}
		if profile.Connection.Provider == freeturn.ProviderDirect {
			directPeer, err = pinDirectPeer(ctx, profile.Connection.Peer, net.DefaultResolver.LookupNetIP)
			if err != nil {
				return nil, err
			}
			profile.Connection.Peer = directPeer
		}
	}

	if mode := profile.Connection.RelayMode; mode != "" && mode != "udp" {
		return nil, errors.New("engine: MVP supports only UDP relay with WG/AWG settings")
	}
	if profile.Connection.WG == nil {
		return nil, errors.New("engine: UDP relay requires WG/AWG settings")
	}
	if err := checkBinary(p.Relay); err != nil {
		return nil, err
	}
	if err := checkBinary(p.Mihomo); err != nil {
		return nil, err
	}
	probe, err := url.Parse(p.ProbeURL)
	if err != nil || probe.Scheme != "https" || probe.Hostname() == "" || probe.User != nil {
		return nil, errors.New("engine: HTTPS probe URL required")
	}
	status := p.ProbeStatus
	if status == 0 {
		status = 204
	}
	if status < 200 || status > 299 {
		return nil, errors.New("engine: invalid probe status")
	}
	dir, err := os.MkdirTemp(p.RuntimeDir, "session-*")
	if err != nil {
		return nil, err
	}
	ok := false
	published := false
	defer func() {
		if !ok {
			_ = os.RemoveAll(dir)
			if published {
				_ = os.Remove(filepath.Join(p.RuntimeDir, "active-session.json"))
			}
		}
	}()
	if err := persistence.WriteJSON(filepath.Join(p.RuntimeDir, "active-session.json"), persistence.SessionManifest{Version: 1, Directory: filepath.Base(dir)}); err != nil {
		return nil, err
	}
	published = true
	relay, mixed, api, err := p.Ports.addresses()
	if p.Addresses != nil {
		relay, mixed, api, err = p.Addresses()
		if err == nil {
			var apiPort int
			apiPort, err = loopbackPort(api)
			if err == nil {
				err = ValidateAddresses(relay, mixed, apiPort)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	args, err := RelayArgs(profile.Connection, relay)
	if err != nil {
		return nil, err
	}
	outbound, err := WireGuard(profile.Connection.WG.Config, relay)
	if err != nil {
		return nil, err
	}
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, err
	}
	secret := hex.EncodeToString(secretBytes)
	mixedPort, _ := loopbackPort(mixed)
	c := config{MixedPort: mixedPort, BindAddress: "127.0.0.1", Mode: "rule", LogLevel: "silent", Controller: api, Secret: secret, Proxies: []Outbound{outbound}, Rules: []string{"MATCH,PROXY"}}
	if runtime.GOOS == "linux" {
		c.LogLevel = "warning"
	}
	if mode == proxymode.Tun {
		c.DNS = tunDNS()
	}
	configPath := filepath.Join(dir, "config.yaml")
	// JSON — подмножество YAML: строки и ключи экранируются стандартным сериализатором.
	if err := persistence.WriteJSON(configPath, c); err != nil {
		return nil, err
	}
	if err := checkRelayContract(ctx, p.Relay.Path, args); err != nil {
		return nil, err
	}
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	check := exec.CommandContext(checkCtx, p.Mihomo.Path, "-t", "-d", dir, "-f", configPath)
	check.WaitDelay = time.Second
	err = check.Run()
	cancel()
	if err != nil {
		return nil, errors.New("mihomo: configuration validation failed")
	}
	plan := &ports.Plan{
		RelayUDPKeepalive: mode == proxymode.Tun && profile.Connection.Provider == freeturn.ProviderVK && profile.Connection.TurnTransport == freeturn.TurnTransportUDP,
		DirectPeer:        directPeer,
		TunReady:          mode == proxymode.Tun,
		Relay:             ports.Command{Path: p.Relay.Path, Args: args, Dir: dir, ClearProxyEnv: true},
		Engine:            ports.Command{CaptureDiagnostics: runtime.GOOS == "linux", Path: p.Mihomo.Path, Args: []string{"-d", dir, "-f", configPath}, Dir: dir}, MixedAddress: mixed,
		Close: func() error { return persistence.CleanupSession(p.RuntimeDir) },
	}
	plan.Ready = func(ctx context.Context) error { return waitMihomo(ctx, api, secret, mixed) }
	plan.MeasureLatency = func(ctx context.Context) (time.Duration, error) {
		start := time.Now()
		err := Probe(ctx, mixed, p.ProbeURL, status)
		return time.Since(start), err
	}
	plan.Probe = func(ctx context.Context) error {
		if err := Probe(ctx, mixed, p.ProbeURL, status); err != nil {
			return err
		}
		if outbound.Type == "wireguard" {
			return ProbeUDP(ctx, mixed)
		}
		return nil
	}
	ok = true
	return plan, nil
}

func waitMihomo(ctx context.Context, api, secret, mixed string) error {
	if err := waitAPI(ctx, api, secret); err != nil {
		return err
	}
	// Mihomo starts its API before applying the configuration and opening
	// inbound listeners. A successful /version alone does not make Probe safe.
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	dialer := net.Dialer{Timeout: time.Second}
	for {
		conn, err := dialer.DialContext(ctx, "tcp", mixed)
		if err == nil {
			conn.Close()
			return nil
		}
		select {
		case <-ctx.Done():
			return errors.New("mihomo: local proxy readiness timed out or cancelled")
		case <-ticker.C:
		}
	}
}

func waitAPI(ctx context.Context, address, secret string) error {
	transport := &http.Transport{Proxy: nil}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		req, err := http.NewRequestWithContext(ctx, "GET", "http://"+address+"/version", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+secret)
		res, err := client.Do(req)
		if err == nil {
			var version struct {
				Version string `json:"version"`
			}
			err = json.NewDecoder(io.LimitReader(res.Body, 4096)).Decode(&version)
			res.Body.Close()
			if res.StatusCode == 200 && err == nil {
				if version.Version != manifest.Version("mihomo") {
					return errors.New("mihomo: unexpected version")
				}
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return errors.New("mihomo: readiness timed out or cancelled")
		case <-ticker.C:
		}
	}
}

// probeError contains only fixed phase labels, never raw transport errors.
type probeError string

func (e probeError) Error() string            { return string(e) }
func (e probeError) PublicDiagnostic() string { return string(e) }

// Probe проверяет TLS и точный статус через mixed inbound; redirects запрещены.
func Probe(ctx context.Context, address, target string, status int) error {
	if _, err := loopbackPort(address); err != nil {
		return err
	}
	u, err := url.Parse(target)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
		return probeError("probe: invalid HTTPS endpoint")
	}
	proxy, _ := url.Parse("http://" + address)
	var connectStatus atomic.Int32
	var phase atomic.Value
	phase.Store("local proxy CONNECT")
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy), TLSHandshakeTimeout: 10 * time.Second,
		OnProxyConnectResponse: func(_ context.Context, _ *url.URL, _ *http.Request, response *http.Response) error {
			connectStatus.Store(int32(response.StatusCode))
			return nil
		},
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		return probeError("probe: request failed")
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
		TLSHandshakeStart: func() { phase.Store("TLS handshake") },
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			if err == nil {
				phase.Store("HTTPS response")
			}
		},
	}))
	response, err := client.Do(req)
	if err != nil {
		return probeDiagnostic(ctx, err, phase.Load().(string), int(connectStatus.Load()))
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode != status {
		return probeError("probe: unexpected HTTP status " + strconv.Itoa(response.StatusCode))
	}
	return nil
}

// Never publish err.Error(): transport errors can contain URLs and credentials.
func probeDiagnostic(ctx context.Context, err error, phase string, connectStatus int) error {
	prefix := "probe: " + phase + ": "
	if connectStatus != 0 && connectStatus != http.StatusOK {
		return probeError("probe: local proxy CONNECT returned HTTP " + strconv.Itoa(connectStatus))
	}
	if ctx.Err() != nil {
		return probeError(prefix + "cancelled or startup deadline exceeded")
	}
	var timeout net.Error
	if errors.As(err, &timeout) && timeout.Timeout() {
		return probeError(prefix + "timed out")
	}
	var certificate *tls.CertificateVerificationError
	if errors.As(err, &certificate) {
		return probeError(prefix + "certificate verification failed")
	}
	var record tls.RecordHeaderError
	if errors.As(err, &record) {
		return probeError(prefix + "received non-TLS response")
	}
	for _, item := range []struct {
		cause error
		label string
	}{
		{io.EOF, "peer closed connection (EOF)"}, {io.ErrUnexpectedEOF, "peer closed connection (unexpected EOF)"},
		{syscall.ECONNRESET, "connection reset"}, {syscall.ECONNREFUSED, "connection refused"},
		{syscall.ENETUNREACH, "network unreachable"}, {syscall.EHOSTUNREACH, "host unreachable"},
		{syscall.EPIPE, "broken pipe"},
	} {
		if errors.Is(err, item.cause) {
			return probeError(prefix + item.label)
		}
	}
	// The concrete error type gives a useful fallback without its private text.
	for errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}
	return probeError(prefix + fmt.Sprintf("transport error (%T)", err))
}
