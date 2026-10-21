package tun

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

// Client owns one lease; losing the controller closes it or lets it expire.
type Client struct {
	PortableHelper string
	helper         *helperProcess
	Socket         string
	RelayPID       func() int
	DirectPeer     func() string
	UDPKeepalive   func() bool
	mu             sync.Mutex
	conn           net.Conn
	session        string
	sequence       uint64
	cancel         context.CancelFunc
	sessionSummary string
	diagnostic     string
	fault          error
	failed         chan struct{}
}

func (c *Client) dial(ctx context.Context) (net.Conn, error) {
	path := c.Socket
	if path == "" {
		path = SocketPath
	}
	return (&net.Dialer{}).DialContext(ctx, "unix", path)
}
func (c *Client) Unavailable(ctx context.Context) string {
	if c.PortableHelper != "" {
		return portableUnavailable(ctx, c.PortableHelper)
	}
	conn, err := c.dial(ctx)
	if err != nil {
		return "TUN helper unavailable; install and start wireglow-tun.service"
	}
	defer conn.Close()
	_, err = exchange(ctx, conn, Request{Version: ProtocolVersion, ID: 1, Operation: "status"})
	if err != nil {
		return err.Error()
	}
	return ""
}
func (c *Client) Apply(ctx context.Context, address string) (result error) {
	port, err := proxyPort(address)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.diagnostic = ""
	c.sessionSummary = ""
	defer func() {
		if result != nil {
			c.diagnostic = result.Error()
		}
	}()
	if c.conn != nil {
		return errors.New("tun: already active")
	}
	if c.RelayPID == nil || c.RelayPID() <= 0 {
		return errors.New("tun: relay process unavailable")
	}
	if c.PortableHelper != "" {
		if err := c.startTemporary(ctx); err != nil {
			return err
		}
	}
	conn, err := c.dial(ctx)
	if err != nil {
		return errors.New("tun: helper unavailable")
	}
	c.sequence++
	request := Request{Version: ProtocolVersion, ID: c.sequence, Operation: "start", ProxyPort: port, RelayPID: c.RelayPID()}
	if c.DirectPeer != nil {
		request.DirectPeer = c.DirectPeer()
	}
	if c.UDPKeepalive != nil {
		request.UDPKeepalive = c.UDPKeepalive()
	}
	response, err := exchange(ctx, conn, request)
	if err != nil {
		conn.Close()
		return err
	}
	c.conn, c.session, c.fault = conn, response.Session, nil
	c.failed = make(chan struct{})
	leaseCtx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go c.heartbeat(leaseCtx, conn)
	return nil
}
func (c *Client) heartbeat(ctx context.Context, conn net.Conn) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		c.mu.Lock()
		if c.conn != conn {
			c.mu.Unlock()
			return
		}
		c.sequence++
		_, err := exchange(ctx, conn, Request{Version: ProtocolVersion, ID: c.sequence, Operation: "heartbeat", Session: c.session})
		if err != nil {
			c.fault = err
			close(c.failed)
			conn.Close()
			c.mu.Unlock()
			return
		}
		c.mu.Unlock()
	}
}
func (c *Client) Recover(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return c.waitTemporary(ctx)
	}
	c.cancel()
	c.sequence++
	_, err := exchange(ctx, c.conn, Request{Version: ProtocolVersion, ID: c.sequence, Operation: "stop", Session: c.session})
	c.conn.Close()
	c.conn = nil
	c.session = ""
	return errors.Join(err, c.fault, c.waitTemporary(ctx))
}

func (c *Client) Failed() <-chan struct{} { c.mu.Lock(); defer c.mu.Unlock(); return c.failed }

func (c *Client) NetworkDiagnostic() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.diagnostic == "" {
		return c.sessionSummary
	}
	if c.sessionSummary == "" {
		return c.diagnostic
	}
	return c.diagnostic + "; " + c.sessionSummary
}
