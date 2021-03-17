package transport

import (
	"context"
	"net"
)

// Dialer dials using a context.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// DialFunc is functional helper for Dialer.
type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

// DialContext implements Dialer.
func (d DialFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d(ctx, network, address)
}
