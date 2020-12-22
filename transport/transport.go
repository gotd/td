package transport

import (
	"context"
	"net"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/proto/codec"
)

// CustomTransport creates transport using user Codec constructor.
func CustomTransport(dialer Dialer, constructor func() Codec) *Transport {
	if dialer == nil {
		dialer = &net.Dialer{}
	}

	return &Transport{
		dialer:      dialer,
		constructor: constructor,
	}
}

// Intermediate creates Intermediate transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#intermediate
func Intermediate(d Dialer) *Transport {
	return CustomTransport(d, func() Codec {
		return codec.Intermediate{}
	})
}

// Full creates Full transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#full
func Full(d Dialer) *Transport {
	return CustomTransport(d, func() Codec {
		return &codec.Full{}
	})
}

// Transport is MTProto connection creator.
type Transport struct {
	dialer      Dialer
	constructor func() Codec
}

// Codec creates new codec using transport settings.
func (t *Transport) Codec() Codec {
	return t.constructor()
}

// Conn is transport connection.
type Conn interface {
	Send(ctx context.Context, b *bin.Buffer) error
	Recv(ctx context.Context, b *bin.Buffer) error
	Close() error
}

// DialContext creates new MTProto connection.
func (t *Transport) DialContext(ctx context.Context, network, address string) (Conn, error) {
	conn, err := t.dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}

	connectionCodec := t.constructor()
	if err := connectionCodec.WriteHeader(conn); err != nil {
		return nil, xerrors.Errorf("write header: %w", err)
	}

	return &connection{
		conn:  conn,
		codec: connectionCodec,
	}, nil
}
