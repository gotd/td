package telegram

import (
	"context"
	"net"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

func (d DialFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d(ctx, network, address)
}
