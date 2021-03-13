package transport

import (
	"context"
	"net"
	"strconv"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/proto/codec"
)

// NewTransport creates transport using user Codec constructor.
func NewTransport(dialer Dialer, getCodec func() Codec) *Transport {
	return &Transport{
		dialer: orDefaultDialer(dialer),
		codec:  getCodec,
	}
}

// Abridged creates Abridged transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#abridged
func Abridged(d Dialer) *Transport {
	return NewTransport(d, func() Codec {
		return codec.Abridged{}
	})
}

// Intermediate creates Intermediate transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#intermediate
func Intermediate(d Dialer) *Transport {
	return NewTransport(d, func() Codec {
		return codec.Intermediate{}
	})
}

// PaddedIntermediate creates Padded intermediate transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#padded-intermediate
func PaddedIntermediate(d Dialer) *Transport {
	return NewTransport(d, func() Codec {
		return codec.PaddedIntermediate{}
	})
}

// Full creates Full transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#full
func Full(d Dialer) *Transport {
	return NewTransport(d, func() Codec {
		return &codec.Full{}
	})
}

// Transport is MTProto connection creator.
type Transport struct {
	dialer Dialer
	codec  func() Codec
}

// Codec creates new codec using transport settings.
func (t *Transport) Codec() Codec {
	return t.codec()
}

// Conn is transport connection.
type Conn interface {
	Send(ctx context.Context, b *bin.Buffer) error
	Recv(ctx context.Context, b *bin.Buffer) error
	Close() error
}

type telegramDialer interface {
	DialTelegram(ctx context.Context, network string, dc int) (Conn, error)
}

func splitAddr(input string, del byte) (dc int, addr string, err error) {
	index := strings.IndexByte(input, del)
	if index < 0 {
		err = xerrors.Errorf("expected delimiter %c in %q", del, input)
		return
	}

	// If del is last character.
	if len(input)-1 == index {
		err = xerrors.Errorf("expected address in %q", input)
		return
	}
	dc, err = strconv.Atoi(input[:index])
	addr = input[index+1:]
	return
}

// DialContext creates new MTProto connection.
func (t *Transport) DialContext(ctx context.Context, network, address string) (Conn, error) {
	dc, addr, err := splitAddr(address, '|')
	if err != nil {
		return nil, xerrors.Errorf("invalid address: %w", err)
	}

	if td, ok := t.dialer.(telegramDialer); ok {
		return td.DialTelegram(ctx, network, dc)
	}

	conn, err := t.dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, xerrors.Errorf("dial: %w", err)
	}

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
