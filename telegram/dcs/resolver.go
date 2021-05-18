package dcs

import (
	"context"
	"net"

	"github.com/gotd/td/transport"
)

var _ Resolver = DefaultResolver()

// Resolver resolves DC and creates transport MTProto connection.
type Resolver interface {
	Primary(ctx context.Context, dc int, list List) (transport.Conn, error)
	MediaOnly(ctx context.Context, dc int, list List) (transport.Conn, error)
	CDN(ctx context.Context, dc int, list List) (transport.Conn, error)
}

// DialFunc connects to the address on the named network.
type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// Protocol is MTProto transport protocol.
//
// See https://core.telegram.org/mtproto/mtproto-transports
type Protocol interface {
	Codec() transport.Codec
	Handshake(conn net.Conn) (transport.Conn, error)
}
