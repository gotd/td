package transport

import (
	"net"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/proto/codec"
)

// NewTransport creates transport using user Codec constructor.
func NewTransport(getCodec func() Codec) *Transport {
	return &Transport{
		codec: getCodec,
	}
}

// Abridged creates Abridged transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#abridged
func Abridged() *Transport {
	return NewTransport(func() Codec {
		return codec.Abridged{}
	})
}

// Intermediate creates Intermediate transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#intermediate
func Intermediate() *Transport {
	return NewTransport(func() Codec {
		return codec.Intermediate{}
	})
}

// PaddedIntermediate creates Padded intermediate transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#padded-intermediate
func PaddedIntermediate() *Transport {
	return NewTransport(func() Codec {
		return codec.PaddedIntermediate{}
	})
}

// Full creates Full transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#full
func Full() *Transport {
	return NewTransport(func() Codec {
		return &codec.Full{}
	})
}

// Transport is MTProto connection creator.
type Transport struct {
	codec func() Codec
}

// Codec creates new codec using transport settings.
func (t *Transport) Codec() Codec {
	return t.codec()
}

// Handshake inits given net.Conn as MTProto connection.
func (t *Transport) Handshake(conn net.Conn) (Conn, error) {
	connCodec := t.codec()
	if err := connCodec.WriteHeader(conn); err != nil {
		return nil, xerrors.Errorf("write header: %w", err)
	}

	return &connection{
		conn:  conn,
		codec: connCodec,
	}, nil
}

// Pipe creates a in-memory MTProto connection.
func (t *Transport) Pipe() (a, b Conn) {
	p1, p2 := net.Pipe()

	return &connection{
			conn:  p1,
			codec: t.codec(),
		}, &connection{
			conn:  p2,
			codec: t.codec(),
		}
}
