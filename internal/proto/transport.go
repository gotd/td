package proto

import (
	"context"
	"net"
	"time"

	"golang.org/x/xerrors"

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

// Intermediate is intermediate MTProto transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#intermediate
type Intermediate struct {
	Dialer Dialer

	conn net.Conn
}

// Dial sends protocol version.
func (i *Intermediate) Dial(ctx context.Context, network, addr string) (err error) {
	i.conn, err = i.Dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return xerrors.Errorf("dial: %w", err)
	}

	if err := i.conn.SetDeadline(deadline(ctx)); err != nil {
		return xerrors.Errorf("set deadline: %w", err)
	}

	if _, err := i.conn.Write(IntermediateClientStart); err != nil {
		return xerrors.Errorf("start intermediate: %w", err)
	}

	if err := i.conn.SetDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("reset connection deadline: %w", err)
	}

	return nil
}

// Send sends message from given buffer.
func (i *Intermediate) Send(ctx context.Context, b *bin.Buffer) error {
	if err := i.conn.SetWriteDeadline(deadline(ctx)); err != nil {
		return xerrors.Errorf("set deadline: %w", err)
	}

	if err := WriteIntermediate(i.conn, b); err != nil {
		return xerrors.Errorf("write intermediate: %w", err)
	}

	if err := i.conn.SetWriteDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("reset connection deadline: %w", err)
	}

	return nil
}

// Recv fills buffer with received message.
func (i *Intermediate) Recv(ctx context.Context, b *bin.Buffer) error {
	if err := i.conn.SetReadDeadline(deadline(ctx)); err != nil {
		return xerrors.Errorf("set deadline: %w", err)
	}

	if err := ReadIntermediate(i.conn, b); err != nil {
		return xerrors.Errorf("read intermediate: %w", err)
	}

	if err := i.conn.SetReadDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("reset connection deadline: %w", err)
	}

	return nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (i *Intermediate) Close() error {
	return i.conn.Close()
}
