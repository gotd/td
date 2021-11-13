package transport

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
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
		return errors.Wrap(err, "reset write deadline")
	}
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.conn.SetWriteDeadline(deadline); err != nil {
			return errors.Wrap(err, "set write deadline")
		}
	}

	if err := c.codec.Write(c.conn, b); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

// Recv reads message to buffer using MTProto connection.
func (c *connection) Recv(ctx context.Context, b *bin.Buffer) error {
	// Serializing access to deadlines.
	c.readMux.Lock()
	defer c.readMux.Unlock()

	if err := c.conn.SetReadDeadline(time.Time{}); err != nil {
		return errors.Wrap(err, "reset read deadline")
	}
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			return errors.Wrap(err, "set read deadline")
		}
	}

	if err := c.codec.Read(c.conn, b); err != nil {
		return errors.Wrap(err, "read")
	}

	return nil
}

// Close closes MTProto connection.
func (c *connection) Close() error {
	return c.conn.Close()
}
