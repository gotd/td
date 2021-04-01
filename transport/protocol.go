package transport

import (
	"net"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/proto/codec"
)

// NewProtocol creates transport protocol using user Codec constructor.
//
// See https://core.telegram.org/mtproto/mtproto-transports
func NewProtocol(getCodec func() Codec) *Protocol {
	return &Protocol{
		codec: getCodec,
	}
}

// Abridged creates Abridged transport protocol.
//
// See https://core.telegram.org/mtproto/mtproto-transports#abridged
func Abridged() *Protocol {
	return NewProtocol(func() Codec {
		return codec.Abridged{}
	})
}

// Intermediate creates Intermediate transport protocol.
//
// See https://core.telegram.org/mtproto/mtproto-transports#intermediate
func Intermediate() *Protocol {
	return NewProtocol(func() Codec {
		return codec.Intermediate{}
	})
}

// PaddedIntermediate creates Padded intermediate transport protocol.
//
// See https://core.telegram.org/mtproto/mtproto-transports#padded-intermediate
func PaddedIntermediate() *Protocol {
	return NewProtocol(func() Codec {
		return codec.PaddedIntermediate{}
	})
}

// Full creates Full transport protocol.
//
// See https://core.telegram.org/mtproto/mtproto-transports#full
func Full() *Protocol {
	return NewProtocol(func() Codec {
		return &codec.Full{}
	})
}

// Protocol is MTProto transport protocol.
//
// See https://core.telegram.org/mtproto/mtproto-transports
type Protocol struct {
	codec func() Codec
}

// Codec creates new codec using protocol settings.
func (t *Protocol) Codec() Codec {
	return t.codec()
}

// Handshake inits given net.Conn as MTProto connection.
func (t *Protocol) Handshake(conn net.Conn) (Conn, error) {
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
func (t *Protocol) Pipe() (a, b Conn) {
	p1, p2 := net.Pipe()

	return &connection{
			conn:  p1,
			codec: t.codec(),
		}, &connection{
			conn:  p2,
			codec: t.codec(),
		}
}
