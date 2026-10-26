package tun

import (
	"context"
	"net"
)

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
