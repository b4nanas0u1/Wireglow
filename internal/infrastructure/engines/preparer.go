package engines

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"wireglow/internal/application/ports"
	"wireglow/internal/domain/profile"
	"wireglow/internal/infrastructure/engines/manifest"
	"wireglow/internal/infrastructure/persistence"
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

type config struct {
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

func (p *Preparer) Prepare(ctx context.Context, profile profile.Profile) (_ *ports.Plan, err error) {
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
		Relay:  ports.Command{Path: p.Relay.Path, Args: args, Dir: dir},
		Engine: ports.Command{Path: p.Mihomo.Path, Args: []string{"-d", dir, "-f", configPath}, Dir: dir}, MixedAddress: mixed,
		Close: func() error { return persistence.CleanupSession(p.RuntimeDir) },
	}
	plan.Ready = func(ctx context.Context) error { return waitAPI(ctx, api, secret) }
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

// Probe проверяет TLS и точный статус через mixed inbound; redirects запрещены.
func Probe(ctx context.Context, address, target string, status int) error {
	if _, err := loopbackPort(address); err != nil {
		return err
	}
	u, err := url.Parse(target)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
		return errors.New("probe: invalid HTTPS endpoint")
	}
	proxy, _ := url.Parse("http://" + address)
	transport := &http.Transport{Proxy: http.ProxyURL(proxy), TLSHandshakeTimeout: 10 * time.Second}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		return errors.New("probe: request failed")
	}
	response, err := client.Do(req)
	if err != nil {
		return errors.New("probe: HTTPS path not confirmed")
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode != status {
		return errors.New("probe: unexpected HTTP status " + strconv.Itoa(response.StatusCode))
	}
	return nil
}
