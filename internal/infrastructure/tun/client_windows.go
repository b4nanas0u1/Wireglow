package tun

import (
	"context"
	"errors"
)

// Reserved for the elevated portable helper (UAC + Job Object). The common
// protocol/config compiler can be reused; activation stays unavailable for now.
type Client struct {
	PortableHelper string
	Socket         string
	RelayPID       func() int
	DirectPeer     func() string
	UDPKeepalive   func() bool
}

func (*Client) Unavailable(context.Context) string { return "TUN is not available on Windows yet" }
func (*Client) Apply(context.Context, string) error {
	return errors.New("tun: Windows helper not implemented")
}
func (*Client) Recover(context.Context) error { return nil }

func (*Client) ConfigurePortable(string) error { return nil }
