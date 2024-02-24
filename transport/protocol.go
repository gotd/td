package transport

import (
	"net"

	"github.com/go-faster/errors"

	"github.com/gotd/td/proto/codec"
)

// Protocol is MTProto transport protocol.
//
// See https://core.telegram.org/mtproto/mtproto-transports
type Protocol struct {
	codec func() Codec
}

// NewProtocol creates new transport protocol using user Codec constructor.
//
// See https://core.telegram.org/mtproto/mtproto-transports
func NewProtocol(getCodec func() Codec) Protocol {
	return Protocol{
		codec: getCodec,
	}
}

// Telegram transport protocols.
//
// See https://core.telegram.org/mtproto/mtproto-transports
var (
	// Abridged is abridged transport protocol.
	//
	// See https://core.telegram.org/mtproto/mtproto-transports#abridged
	Abridged = NewProtocol(func() Codec { return codec.Abridged{} })

	// Intermediate is intermediate transport protocol.
	//
	// See https://core.telegram.org/mtproto/mtproto-transports#intermediate
	Intermediate = NewProtocol(func() Codec { return codec.Intermediate{} })

	// PaddedIntermediate is padded intermediate transport protocol.
	//
	// See https://core.telegram.org/mtproto/mtproto-transports#padded-intermediate
	PaddedIntermediate = NewProtocol(func() Codec { return codec.PaddedIntermediate{} })

	// Full is full transport protocol.
	//
	// See https://core.telegram.org/mtproto/mtproto-transports#full
	Full = NewProtocol(func() Codec { return &codec.Full{} })
)

// Codec creates new codec using protocol settings.
func (p Protocol) Codec() Codec {
	return p.codec()
}

// CodecNoHeader is Codec without header.
func (p Protocol) CodecNoHeader() Codec {
	return codec.NoHeader{Codec: p.codec()}
}

// Handshake inits given net.Conn as MTProto connection.
func (p Protocol) Handshake(conn net.Conn) (Conn, error) {
	connCodec := p.codec()
	if err := connCodec.WriteHeader(conn); err != nil {
		return nil, errors.Wrap(err, "write header")
	}

	return &connection{
		conn:  conn,
		codec: connCodec,
	}, nil
}

// Pipe creates an in-memory MTProto connection.
func (p Protocol) Pipe() (a, b Conn) {
	p1, p2 := net.Pipe()

	return &connection{
			conn:  p1,
			codec: p.codec(),
		}, &connection{
			conn:  p2,
			codec: p.codec(),
		}
}
