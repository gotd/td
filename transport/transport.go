package transport

import (
	"context"
	"net"

	"golang.org/x/xerrors"

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

// DialContext creates new MTProto connection.
func (t *Transport) DialContext(ctx context.Context, network, address string) (Connection, error) {
	conn, err := t.dialer.DialContext(ctx, network, address)
	if err != nil {
		return Connection{}, xerrors.Errorf("dial: %w", err)
	}

	connectionCodec := t.constructor()
	if err := connectionCodec.WriteHeader(conn); err != nil {
		return Connection{}, err
	}

	return Connection{
		conn:  conn,
		codec: connectionCodec,
	}, nil
}
