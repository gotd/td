package transport

import (
	"context"
	"net"
	"sync"
	"time"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
)

// Conn is transport connection.
type Conn interface {
	Send(ctx context.Context, b *bin.Buffer) error
	Recv(ctx context.Context, b *bin.Buffer) error
	Close() error
}

var _ Conn = (*connection)(nil)

// connection is MTProto connection.
type connection struct {
	conn  net.Conn
	codec Codec

	readMux  sync.Mutex
	writeMux sync.Mutex
}

// Send sends message from buffer using MTProto connection.
func (c *connection) Send(ctx context.Context, b *bin.Buffer) error {
	// Serializing access to deadlines.
	c.writeMux.Lock()
	defer c.writeMux.Unlock()

	if err := c.conn.SetWriteDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("reset write deadline: %w", err)
	}
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.conn.SetWriteDeadline(deadline); err != nil {
			return xerrors.Errorf("set write deadline: %w", err)
		}
	}

	if err := c.codec.Write(c.conn, b); err != nil {
		return xerrors.Errorf("write: %w", err)
	}

	return nil
}

// Recv reads message to buffer using MTProto connection.
func (c *connection) Recv(ctx context.Context, b *bin.Buffer) error {
	// Serializing access to deadlines.
	c.readMux.Lock()
	defer c.readMux.Unlock()

	if err := c.conn.SetReadDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("reset read deadline: %w", err)
	}
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			return xerrors.Errorf("set read deadline: %w", err)
		}
	}

	if err := c.codec.Read(c.conn, b); err != nil {
		return xerrors.Errorf("read: %w", err)
	}

	return nil
}

// Close closes MTProto connection.
func (c *connection) Close() error {
	return c.conn.Close()
}
