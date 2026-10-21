package tun

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"wireglow/internal/infrastructure/engines/manifest"

	"golang.org/x/sys/unix"
)

// Server is installed root-owned together with its fixed mihomo binary.
// It never executes a path received over IPC. Runtime must be root-only.
type Server struct {
	udpKeeper     *udpKeepalive
	relayPID      int
	relayRoutes   []relayRoute
	relayUDP      []relayUDPRoute
	OwnerPID      int
	OneShot       bool
	attempted     atomic.Bool
	networkLock   *os.File
	cleanupFailed bool
	UID           uint32
	Binary        string
	Runtime       string
	Socket        string
	mu            sync.Mutex
}
type journal struct {
	Direct   []directRoute   `json:"direct,omitempty"`
	RelayUDP []relayUDPRoute `json:"relayUDP,omitempty"`
	RelayTCP []relayRoute    `json:"relayTCP,omitempty"`
	Device   string          `json:"device"`
	Uplink   string          `json:"uplink"`
}

func protected(path string) error {
	path = filepath.Clean(path)
	for {
		s, err := os.Lstat(path)
		if err != nil {
			return err
		}
		st, ok := s.Sys().(*syscall.Stat_t)
		if !ok || st.Uid != 0 || s.Mode()&0022 != 0 || s.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("tun: helper path must be root-owned, non-writable and without symlinks: %s", path)
		}
		parent := filepath.Dir(path)
		if parent == path {
			return nil
		}
		path = parent
	}
}
func ip(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/bin/ip", args...)
	cmd.Env = []string{"PATH=/usr/sbin:/usr/bin", "LC_ALL=C"}
	b, err := cmd.Output()
	if err != nil {
		detail := err.Error()
		var exit *exec.ExitError
		if errors.As(err, &exit) && len(exit.Stderr) > 0 {
			message := exit.Stderr
			if len(message) > 512 {
				message = message[:512]
			}
			detail = strings.TrimSpace(string(message))
		}
		return nil, fmt.Errorf("tun: ip %s failed: %s", strings.Join(args, " "), detail)
	}
	return b, nil
}

type tunnelLink struct {
	Name  string   `json:"ifname"`
	Type  string   `json:"link_type"`
	Flags []string `json:"flags"`
	State string   `json:"operstate"`
	Info  struct {
		Kind string `json:"info_kind"`
	} `json:"linkinfo"`
}

func (l tunnelLink) activeTunnel() bool {
	if l.Type != "none" && l.Info.Kind != "wireguard" && l.Info.Kind != "tun" {
		return false
	}
	up := false
	for _, flag := range l.Flags {
		if flag == "NO-CARRIER" {
			return false
		}
		if flag == "UP" {
			up = true
		}
	}
	// Persistent TUN devices can remain administratively UP after their owner
	// disconnects. UNKNOWN is normal for a working TUN/WireGuard interface.
	return up && l.State != "DOWN"
}

func uplink(ctx context.Context, ownedDevice string) (string, error) {
	b, err := ip(ctx, "-j", "link", "show")
	if err != nil {
		return "", err
	}
	var links []tunnelLink
	if err = json.Unmarshal(b, &links); err != nil {
		return "", err
	}
	for _, l := range links {
		if l.Name != ownedDevice && l.activeTunnel() {
			return "", fmt.Errorf("tun: another VPN/tunnel is active (%s); TUN activation refused", l.Name)
		}
	}
	b, err = ip(ctx, "-j", "route", "get", BootstrapDNS)
	if err != nil {
		return "", err
	}
	var routes []struct {
		Dev string `json:"dev"`
	}
	if err = json.Unmarshal(b, &routes); err != nil || len(routes) != 1 || routes[0].Dev == "" || routes[0].Dev == "lo" {
		return "", errors.New("tun: cannot determine uplink")
	}
	return routes[0].Dev, nil
}
func (s *Server) preflight(ctx context.Context) (string, error) {
	device, err := uplink(ctx, "")
	if err != nil {
		return "", err
	}
	var b []byte
	for _, family := range []string{"-4", "-6"} {
		b, err = ip(ctx, family, "-j", "rule", "show")
		if err != nil {
			return "", err
		}
		var rules []struct {
			Priority int `json:"priority"`
			Table    any `json:"table"`
		}
		if err = json.Unmarshal(b, &rules); err != nil {
			return "", err
		}
		for _, r := range rules {
			if r.Priority == 10216 || r.Priority == 10217 || r.Priority == 10218 || r.Priority == 10219 || r.Priority == 10220 || fmt.Sprint(r.Table) == "20220" {
				return "", errors.New("tun: reserved routing resources are busy")
			}
		}
		b, err = ip(ctx, family, "-j", "route", "show", "table", "all")
		if err != nil {
			return "", err
		}
		var routes []struct {
			Table any `json:"table"`
		}
		if err = json.Unmarshal(b, &routes); err != nil {
			return "", err
		}
		for _, r := range routes {
			if fmt.Sprint(r.Table) == "20220" {
				return "", errors.New("tun: reserved routing table is busy")
			}
		}
	}
	return device, nil
}
func (s *Server) checkUplink(ctx context.Context) error {
	b, err := os.ReadFile(filepath.Join(s.Runtime, "journal.json"))
	if err != nil {
		return err
	}
	var j journal
	if err = json.Unmarshal(b, &j); err != nil {
		return err
	}
	device, err := uplink(ctx, j.Device)
	if err != nil {
		return err
	}
	if device != j.Uplink {
		return errors.New("tun: uplink changed; reconnect required")
	}
	return nil
}

func (s *Server) writeJournal(j journal) error {
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	path := filepath.Join(s.Runtime, "journal.json")
	f, err := os.OpenFile(path+".tmp", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	err = errors.Join(err, f.Sync(), f.Close())
	if err != nil {
		return err
	}
	return os.Rename(path+".tmp", path)
}

// Cleanup deletes exact owned rules only. No defaults, resolver settings or
// firewall rules are modified; removing the owned link removes its routes.
func (s *Server) Cleanup(ctx context.Context) (result error) {
	defer func() { s.cleanupFailed = result != nil }()
	path := filepath.Join(s.Runtime, "journal.json")
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var j journal
	if json.Unmarshal(b, &j) != nil || len(j.Device) != 14 || !strings.HasPrefix(j.Device, "wg") {
		return errors.New("tun: invalid recovery journal")
	}
	for _, route := range j.RelayUDP {
		if !route.valid() {
			return errors.New("tun: invalid UDP recovery route")
		}
	}
	if len(j.Direct) > 1 {
		return errors.New("tun: invalid direct recovery routes")
	}
	for _, route := range j.Direct {
		if !route.valid() {
			return errors.New("tun: invalid direct recovery route")
		}
	}
	for _, route := range j.RelayTCP {
		if !route.valid() {
			return errors.New("tun: invalid relay recovery route")
		}
	}
	// Read rules and delete only those identical to the recorded recipe. Missing
	// resources are success, while an unexpected occupant keeps the journal.
	for _, family := range []string{"-4", "-6"} {
		b, e := ip(ctx, family, "-j", "rule", "show")
		if e != nil {
			return e
		}
		var rules []struct {
			UIDStart *uint32         `json:"uid_start"`
			UIDEnd   *uint32         `json:"uid_end"`
			Src      string          `json:"src"`
			Protocol string          `json:"ipproto"`
			Sport    any             `json:"sport"`
			Dport    any             `json:"dport"`
			Priority int             `json:"priority"`
			Table    any             `json:"table"`
			Dst      string          `json:"dst"`
			Mark     any             `json:"fwmark"`
			Not      json.RawMessage `json:"not"`
			OIF      string          `json:"oif"`
		}
		if e = json.Unmarshal(b, &rules); e != nil {
			return e
		}
		for _, r := range rules {
			if r.Priority == 10216 {
				matched := false
				for _, route := range j.Direct {
					args := route.args("del")
					host, port, _ := net.SplitHostPort(route.Peer)
					if args[0] == family && r.Src == "all" && r.Dst == host && r.Mark == nil && len(r.Not) == 0 && r.OIF == "" && r.UIDStart != nil && r.UIDEnd != nil && *r.UIDStart == route.UID && *r.UIDEnd == route.UID && r.Protocol == "udp" && r.Sport == nil && fmt.Sprint(r.Dport) == port && fmt.Sprint(r.Table) == "main" {
						matched = true
						if _, e = ip(ctx, args...); e != nil {
							return e
						}
						break
					}
				}
				if !matched {
					return errors.New("tun: direct route ownership changed; recovery journal retained")
				}
				continue
			}
			if r.Priority == 10217 {
				matched := false
				for _, route := range j.RelayUDP {
					if route.Family == family && r.Src == "all" && r.Dst == "" && r.Mark == nil && len(r.Not) == 0 && r.OIF == "" && r.UIDStart != nil && r.UIDEnd != nil && *r.UIDStart == route.UID && *r.UIDEnd == route.UID && r.Protocol == "udp" && fmt.Sprint(r.Sport) == strconv.Itoa(int(route.Port)) && r.Dport == nil && fmt.Sprint(r.Table) == "main" {
						matched = true
						if _, e = ip(ctx, route.args("del")...); e != nil {
							return e
						}
						break
					}
				}
				if !matched {
					return errors.New("tun: UDP route ownership changed; recovery journal retained")
				}
				continue
			}
			if r.Priority == 10218 {
				matched := false
				for _, route := range j.RelayTCP {
					if route.Family == family && r.Src == route.Local && r.Dst == route.Remote && r.Protocol == "tcp" && fmt.Sprint(r.Sport) == strconv.Itoa(int(route.LocalPort)) && fmt.Sprint(r.Dport) == strconv.Itoa(int(route.RemotePort)) && fmt.Sprint(r.Table) == "main" {
						matched = true
						if _, e = ip(ctx, route.args("del")...); e != nil {
							return e
						}
						break
					}
				}
				if !matched {
					return errors.New("tun: relay route ownership changed; recovery journal retained")
				}
				continue
			}
			// sing-tun creates an IPv6 output-interface rule even with
			// auto-route disabled. Its exact device/table pair is session-owned.
			if family == "-6" && r.OIF == j.Device && fmt.Sprint(r.Table) == "20220" {
				if _, e = ip(ctx, family, "rule", "del", "pref", strconv.Itoa(r.Priority), "oif", j.Device, "lookup", "20220"); e != nil {
					return e
				}
				continue
			}
			if r.Priority != 10219 && r.Priority != 10220 {
				continue
			}
			var args []string
			if r.Priority == 10219 && family == "-4" && r.Dst == BootstrapDNS && fmt.Sprint(r.Table) == "main" {
				args = []string{family, "rule", "del", "pref", "10219", "to", BootstrapDNS, "lookup", "main"}
			} else if r.Priority == 10220 && fmt.Sprint(r.Table) == "20220" && len(r.Not) > 0 && (fmt.Sprint(r.Mark) == "0x4efc" || fmt.Sprint(r.Mark) == "20220") {
				args = []string{family, "rule", "del", "pref", "10220", "not", "fwmark", "20220", "lookup", "20220"}
			} else {
				return errors.New("tun: routing ownership changed; recovery journal retained")
			}
			if _, e = ip(ctx, args...); e != nil {
				return e
			}
		}
	}
	if _, err = net.InterfaceByName(j.Device); err == nil {
		if _, err = ip(ctx, "link", "del", j.Device); err != nil {
			return err
		}
	}
	return os.Remove(path)
}
func checkEngineDigest(path string) error {
	expected, err := manifest.Field("mihomo.platforms.linux-amd64.binary-sha256")
	if err != nil || expected == "" {
		return errors.New("tun: pinned engine digest unavailable")
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	digest := sha256.New()
	if _, err = io.Copy(digest, file); err != nil {
		return err
	}
	if hex.EncodeToString(digest.Sum(nil)) != expected {
		return errors.New("tun: privileged engine digest mismatch")
	}
	return nil
}

// All helper deployments share one network lease and journal, while each
// resident service / temporary owner has an independent listener lifetime.
var errNetworkBusy = errors.New("tun: another helper owns the TUN session")

func (s *Server) acquireNetwork() error {
	if s.networkLock != nil {
		return errors.New("tun: network session already active")
	}
	fd, err := unix.Open(filepath.Join(s.Runtime, "session.lock"), unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return err
	}
	file := os.NewFile(uintptr(fd), "tun-network-lease")
	if err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		file.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return errNetworkBusy
		}
		return err
	}
	s.networkLock = file
	return nil
}
func (s *Server) releaseNetwork() {
	if s.networkLock != nil {
		_ = s.networkLock.Close()
		s.networkLock = nil
	}
}
func (s *Server) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if os.Geteuid() != 0 {
		return errors.New("tun: helper requires root")
	}
	if !filepath.IsAbs(s.Binary) || !filepath.IsAbs(s.Runtime) || !filepath.IsAbs(s.Socket) {
		return errors.New("tun: absolute installation paths required")
	}
	if err := protected(s.Binary); err != nil {
		return err
	}
	if err := checkEngineDigest(s.Binary); err != nil {
		return err
	}
	if err := protected("/usr/bin/ip"); err != nil {
		return err
	}
	if err := os.MkdirAll(s.Runtime, 0700); err != nil {
		return err
	}
	if err := protected(s.Runtime); err != nil {
		return err
	}
	instanceLock := "helper.lock"
	if s.OwnerPID != 0 {
		instanceLock = "helper-" + strconv.Itoa(s.OwnerPID) + ".lock"
	}
	lock, err := os.OpenFile(filepath.Join(s.Runtime, instanceLock), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err = unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		return errors.New("tun: helper already running")
	}
	if err = s.acquireNetwork(); err == nil {
		recoveryErr := s.Cleanup(ctx)
		s.releaseNetwork()
		if recoveryErr != nil {
			return recoveryErr
		}
	} else if !errors.Is(err, errNetworkBusy) {
		return err
	}
	_ = os.Remove(s.Socket)
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: s.Socket, Net: "unix"})
	if err != nil {
		return err
	}
	defer listener.Close()
	if err = os.Chmod(s.Socket, 0600); err != nil {
		return err
	}
	if err = os.Chown(s.Socket, int(s.UID), -1); err != nil {
		return err
	}
	if s.OneShot {
		go func() {
			timer := time.NewTimer(20 * time.Second)
			defer timer.Stop()
			select {
			case <-ctx.Done():
			case <-timer.C:
				if !s.attempted.Load() {
					cancel()
				}
			}
		}()
	}
	var workers sync.WaitGroup
	defer workers.Wait()
	go func() { <-ctx.Done(); listener.Close() }()
	for {
		conn, e := listener.AcceptUnix()
		if e != nil {
			if ctx.Err() != nil {
				return nil
			}
			return e
		}
		workers.Add(1)
		go func() {
			defer workers.Done()
			defer conn.Close()
			if s.handle(ctx, conn) && s.OneShot {
				cancel()
			}
		}()
	}
}
func peer(conn *net.UnixConn) (*unix.Ucred, error) {
	raw, err := conn.SyscallConn()
	if err != nil {
		return nil, err
	}
	var cred *unix.Ucred
	var inner error
	err = raw.Control(func(fd uintptr) { cred, inner = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED) })
	if err != nil {
		return nil, err
	}
	if inner != nil {
		return nil, inner
	}
	return cred, nil
}
func (s *Server) handle(ctx context.Context, conn *net.UnixConn) (attempted bool) {
	cred, err := peer(conn)
	if err != nil || cred.Uid != s.UID || (s.OwnerPID != 0 && int(cred.Pid) != s.OwnerPID) {
		return
	}
	uid := cred.Uid
	var session string
	var lastID uint64
	var cmd *exec.Cmd
	var exited chan error
	defer func() {
		if cmd != nil {
			s.stop(cmd, exited)
		}
		if session != "" {
			s.mu.Unlock()
		}
	}()
	for {
		conn.SetReadDeadline(time.Now().Add(Lease))
		var req Request
		if readMessage(conn, &req) != nil {
			return
		}
		response := Response{Version: ProtocolVersion, ID: req.ID, Session: session}
		if req.Version != ProtocolVersion || req.ID <= lastID {
			response.Error = "tun: protocol mismatch"
		} else {
			lastID = req.ID
			switch req.Operation {
			case "status":
				if !s.mu.TryLock() {
					response.Error = "tun: another session is active"
				} else {
					_, e := s.preflight(ctx)
					if e != nil {
						response.Error = e.Error()
					}
					s.mu.Unlock()
				}
			case "start":
				attempted = true
				s.attempted.Store(true)
				if session != "" {
					response.Error = "tun: already started"
				} else if !s.mu.TryLock() {
					response.Error = "tun: another session is active"
				} else {
					var e error
					cmd, exited, session, e = s.start(ctx, uid, req)
					response.Session = session
					if e != nil {
						response.Error = e.Error()
						s.mu.Unlock()
					}
				}
			case "heartbeat":
				if session == "" || req.Session != session {
					response.Error = "tun: invalid session"
				} else {
					select {
					case <-ctx.Done():
						response.Error = "tun: helper stopping"
					case <-exited:
						response.Error = "tun: engine exited"
					default:
						if e := s.checkUplink(ctx); e != nil {
							response.Error = e.Error()
						} else if s.udpKeeper != nil {
							if e := s.udpKeeper.failure(); e != nil {
								response.Error = e.Error()
							}
						}
					}
				}
			case "cancel":
				// A cancelled activation may lose the socket before sending Start. Only
				// the bound portable owner may terminate an idle one-shot helper.
				if s.OneShot && session == "" && s.mu.TryLock() {
					s.mu.Unlock()
					attempted = true
					s.attempted.Store(true)
				} else {
					response.Error = "tun: helper is not idle and temporary"
				}
			case "stop":
				if session == "" || req.Session != session {
					response.Error = "tun: invalid session"
				} else {
					err := s.stop(cmd, exited)
					cmd = nil
					s.mu.Unlock()
					session = ""
					if err != nil {
						response.Error = err.Error()
					}
				}
			default:
				response.Error = "tun: unsupported operation"
			}
		}
		conn.SetWriteDeadline(time.Now().Add(time.Second))
		if json.NewEncoder(conn).Encode(response) != nil || req.Operation == "stop" || req.Operation == "cancel" || response.Error != "" {
			return
		}
	}
}
func (s *Server) stop(cmd *exec.Cmd, exited chan error) error {
	if s.networkLock == nil {
		return nil
	}
	defer s.releaseNetwork()
	if s.udpKeeper != nil {
		s.udpKeeper.close()
		s.udpKeeper = nil
	}
	if s.relayPID > 0 {
		current, err := relayTCPRoutes(s.relayPID)
		if err == nil {
			present := 0
			for _, old := range s.relayRoutes {
				for _, now := range current {
					if old == now {
						present++
						break
					}
				}
			}
			fmt.Fprintf(os.Stderr, "wireglow-tun-udp-summary: UDP bypasses=%d\n", len(s.relayUDP))
			fmt.Fprintf(os.Stderr, "wireglow-tun-summary: TCP bypasses=%d; sockets present=%d\n", len(s.relayRoutes), present)
		}
	}
	_ = cmd.Process.Signal(unix.SIGTERM)
	select {
	case <-exited:
	case <-time.After(3 * time.Second):
		_ = cmd.Process.Kill()
		<-exited
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.Cleanup(ctx)
	s.cleanupFailed = err != nil
	return err
}
func (s *Server) start(ctx context.Context, uid uint32, req Request) (*exec.Cmd, chan error, string, error) {
	if err := s.acquireNetwork(); err != nil {
		return nil, nil, "", err
	}
	keepLease := false
	defer func() {
		if !keepLease {
			s.releaseNetwork()
		}
	}()
	if err := checkEngineDigest(s.Binary); err != nil {
		return nil, nil, "", err
	}
	if req.ProxyPort < 1 || req.ProxyPort > 65535 || req.RelayPID <= 0 || (req.UDPKeepalive && req.DirectPeer != "") {
		return nil, nil, "", errors.New("tun: invalid start request")
	}
	stat, err := os.Stat("/proc/" + strconv.Itoa(req.RelayPID))
	if err != nil {
		return nil, nil, "", errors.New("tun: relay not found")
	}
	owner, ok := stat.Sys().(*syscall.Stat_t)
	if !ok || owner.Uid != uid {
		return nil, nil, "", errors.New("tun: relay owner mismatch")
	}
	path, err := os.Readlink("/proc/" + strconv.Itoa(req.RelayPID) + "/exe")
	if err != nil {
		return nil, nil, "", err
	}
	if filepath.Base(path) != "free-turn-proxy" {
		return nil, nil, "", errors.New("tun: unexpected relay executable")
	}
	var keeper *udpKeepalive
	if req.UDPKeepalive {
		keeper, err = newUDPKeepalive(req.RelayPID, uid)
		if err != nil {
			return nil, nil, "", err
		}
		defer func() {
			if !keepLease {
				keeper.close()
			}
		}()
	}
	if err := s.Cleanup(ctx); err != nil {
		return nil, nil, "", err
	}
	uplink, err := s.preflight(ctx)
	if err != nil {
		return nil, nil, "", err
	}
	token := make([]byte, 6)
	if _, err = rand.Read(token); err != nil {
		return nil, nil, "", err
	}
	session := hex.EncodeToString(token)
	device := "wg" + session
	b, err := Compile(req.ProxyPort, path, uplink, device)
	if err != nil {
		return nil, nil, "", err
	}
	config := filepath.Join(s.Runtime, "config.json")
	if err = os.WriteFile(config, b, 0600); err != nil {
		return nil, nil, "", err
	}
	relayRoutes, err := relayTCPRoutes(req.RelayPID)
	if err != nil {
		return nil, nil, "", err
	}
	direct, err := directRoutes(req.DirectPeer, uid)
	if err != nil {
		return nil, nil, "", err
	}
	var udpRoutes []relayUDPRoute
	if req.DirectPeer == "" {
		udpRoutes, err = relayUDPRoutes(req.RelayPID, uid)
		if err != nil {
			return nil, nil, "", err
		}
	}
	if err = s.writeJournal(journal{Device: device, Uplink: uplink, RelayTCP: relayRoutes, RelayUDP: udpRoutes, Direct: direct}); err != nil {
		return nil, nil, "", err
	}
	cmd := exec.Command(s.Binary, "-d", s.Runtime, "-f", config)
	cmd.Env = []string{"PATH=/usr/sbin:/usr/bin", "HOME=" + s.Runtime, "SKIP_SYSTEM_IPV6_CHECK=1"}
	cmd.SysProcAttr = &unix.SysProcAttr{Pdeathsig: unix.SIGKILL}
	if err = cmd.Start(); err != nil {
		_ = s.Cleanup(ctx)
		return nil, nil, "", errors.New("tun: engine start failed")
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait(); close(exited) }()
	fail := func(e error) (*exec.Cmd, chan error, string, error) {
		cleanup := s.stop(cmd, exited)
		return nil, nil, "", errors.Join(e, cleanup)
	}
	deadline := time.NewTimer(8 * time.Second)
	defer deadline.Stop()
	ready := time.NewTicker(50 * time.Millisecond)
	defer ready.Stop()
	for {
		select {
		case <-ctx.Done():
			return fail(ctx.Err())
		case <-exited:
			return fail(errors.New("tun: engine exited during startup"))
		case <-deadline.C:
			return fail(errors.New("tun: device readiness timed out"))
		case <-ready.C:
		}
		if _, err = net.InterfaceByName(device); err == nil {
			break
		}
	}
	for _, route := range direct {
		if _, err = ip(ctx, route.args("add")...); err != nil {
			return fail(err)
		}
	}
	for _, route := range udpRoutes {
		if _, err = ip(ctx, route.args("add")...); err != nil {
			return fail(err)
		}
	}
	for _, route := range relayRoutes {
		if _, err = ip(ctx, route.args("add")...); err != nil {
			return fail(err)
		}
	}
	// All changes have a write-ahead journal and use a dedicated table/rule pair.
	for _, family := range []string{"-4", "-6"} {
		routes, e := ip(ctx, family, "-j", "route", "show", "table", "all")
		if e != nil {
			return fail(e)
		}
		var entries []struct {
			Dst   string `json:"dst"`
			Dev   string `json:"dev"`
			Table any    `json:"table"`
		}
		if e = json.Unmarshal(routes, &entries); e != nil {
			return fail(e)
		}
		exists := false
		for _, entry := range entries {
			if fmt.Sprint(entry.Table) != "20220" {
				continue
			}
			if entry.Dev != device {
				return fail(errors.New("tun: routing table ownership changed"))
			}
			if entry.Dst == "default" {
				exists = true
			}
		}
		if !exists {
			if _, e = ip(ctx, family, "route", "add", "default", "dev", device, "table", "20220"); e != nil {
				return fail(e)
			}
		}
	}
	commands := [][]string{
		{"-4", "rule", "add", "pref", "10219", "to", BootstrapDNS, "lookup", "main"},
		{"-4", "rule", "add", "pref", "10220", "not", "fwmark", "20220", "lookup", "20220"},
		{"-6", "rule", "add", "pref", "10220", "not", "fwmark", "20220", "lookup", "20220"},
	}
	for _, args := range commands {
		if _, err = ip(ctx, args...); err != nil {
			return fail(err)
		}
	}
	keepLease = true
	if keeper != nil {
		s.udpKeeper = keeper
		keeper.start()
	}
	s.relayPID = req.RelayPID
	s.relayRoutes = relayRoutes
	s.relayUDP = udpRoutes
	return cmd, exited, session, nil
}
