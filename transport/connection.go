package transport

import (
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"
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

var (
	_ Conn        = (*connection)(nil)
	_ ReadCounter = (*connection)(nil)
)

// connection is MTProto connection.
type connection struct {
	conn  net.Conn
	codec Codec

	readMux  sync.Mutex
	writeMux sync.Mutex

	// bytesRead counts bytes read from the underlying connection, including
	// bytes of a frame that is not yet complete.
	bytesRead atomic.Uint64
	// reader is created on the first Recv and reused so the io.Reader passed
	// to the codec does not allocate on every read. Constructors leave it nil;
	// the atomic counter above is valid at its zero value.
	reader *countingReader
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

	if c.reader == nil {
		c.reader = &countingReader{r: c.conn, n: &c.bytesRead}
	}
	if err := c.codec.Read(c.reader, b); err != nil {
		return errors.Wrap(err, "read")
	}

	return nil
}

// BytesRead returns total bytes read from the underlying connection.
func (c *connection) BytesRead() uint64 { return c.bytesRead.Load() }

// countingReader counts successful reads into n.
//
// A positive read is progress even when the codec has not finished a frame,
// matching TDLib last_read_at_ (any socket read of more than zero bytes).
type countingReader struct {
	r io.Reader
	n *atomic.Uint64
}

func (r countingReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if n > 0 {
		r.n.Add(uint64(n))
	}
	return n, err
}

// Close closes MTProto connection.
func (c *connection) Close() error {
	return c.conn.Close()
}
