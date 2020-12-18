package proto

import (
	"context"
	"net"

	"github.com/gotd/td/bin"
)

// Dialer dials using a context.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Transport is MTProto transport protocol abstraction.
type Transport interface {
	// Dial sends protocol version.
	Dial(ctx context.Context, network, addr string) error

	// Send sends message from given buffer.
	Send(ctx context.Context, b *bin.Buffer) error

	// Recv fills buffer with received message.
	Recv(ctx context.Context, b *bin.Buffer) error

	// Close closes the connection.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error
}

// check that Intermediate implements Transport in compile time.
var _ Transport = &Intermediate{}
