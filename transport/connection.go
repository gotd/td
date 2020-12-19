package transport

import (
	"context"
	"net"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
)

// Connection is MTProto connection.
type Connection struct {
	conn  net.Conn
	codec Codec
}

// Send sends message from buffer using MTProto connection.
func (c Connection) Send(ctx context.Context, b *bin.Buffer) error {
	if err := c.conn.SetWriteDeadline(deadline(ctx)); err != nil {
		return xerrors.Errorf("set write deadline: %w", err)
	}

	if err := c.codec.Write(c.conn, b); err != nil {
		return xerrors.Errorf("send: %w", err)
	}

	if err := c.conn.SetWriteDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("reset write deadline: %w", err)
	}

	return nil
}

// Recv reads message to buffer using MTProto connection.
func (c Connection) Recv(ctx context.Context, b *bin.Buffer) error {
	if err := c.conn.SetReadDeadline(deadline(ctx)); err != nil {
		return xerrors.Errorf("set read deadline: %w", err)
	}

	if err := c.codec.Read(c.conn, b); err != nil {
		return xerrors.Errorf("recv: %w", err)
	}

	if err := c.conn.SetReadDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("reset read deadline: %w", err)
	}

	return nil
}

// Close closes MTProto connection.
func (c Connection) Close() error {
	return c.conn.Close()
}
