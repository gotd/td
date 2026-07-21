package transport

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

// ErrWriteFailed is returned when writing to the underlying connection failed.
//
// A request that failed with this error is safe to retry on a new connection:
// the message was either not sent at all, or sent partially, and a partial
// MTProto frame is discarded by the server.
var ErrWriteFailed = errors.New("transport write failed")

// writeError annotates a write failure with ErrWriteFailed.
//
// A hand-rolled type is used instead of errors.Join(ErrWriteFailed, cause):
// errors.Join renders as "transport write failed\n<cause>", spanning two lines
// and breaking grep and line-oriented log ingestion — which matters directly
// for the production debugging this transport is meant to serve. rpc/errors.go
// hand-rolls its error types for the same reason.
//
// Unwrap exposes the cause, so errors.Is against the underlying error (notably
// net.ErrClosed) keeps working; Is additionally matches the ErrWriteFailed
// sentinel that callers classify retryability on.
type writeError struct {
	cause error
}

// writeFailed marks err as a write failure. cause must be non-nil.
func writeFailed(cause error) error {
	return &writeError{cause: cause}
}

func (e *writeError) Error() string {
	return ErrWriteFailed.Error() + ": " + e.cause.Error()
}

func (e *writeError) Unwrap() error { return e.cause }

func (e *writeError) Is(target error) bool { return target == ErrWriteFailed }

// forcedDeadline is a deadline in the past, used to unblock pending I/O.
var forcedDeadline = time.Unix(1, 0)

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

	// readSem and writeSem serialize access to the connection and its
	// deadlines. They are capacity-1 channels rather than sync.Mutex so that
	// acquisition can be cancelled: a peer that stops reading can park a Write
	// indefinitely, and with a plain mutex every other sender — including the
	// ping loop, which is the only liveness detector — would be trapped behind
	// it with no way to observe its own timeout.
	readSem  chan struct{}
	writeSem chan struct{}

	// deadlineMu serializes the closed check against the deadline calls in
	// Close/Send/Recv. It is held only across those non-blocking calls, never
	// across codec.Write/codec.Read: without it, Close could run its
	// closed=true+SetDeadline(forcedDeadline) in the window between Send's
	// closed check and Send's own SetDeadline(zero), erasing the forced past
	// deadline and parking the I/O forever on a net.Conn that does not
	// unblock on Close by itself. Because nothing ever holds deadlineMu
	// across blocking I/O, Close can never block on it.
	deadlineMu sync.Mutex

	// closed is set before Close forces the deadlines into the past. Send
	// and Recv check it (under deadlineMu) right after acquiring their
	// semaphore, before touching any deadline. A plain bool, not
	// atomic.Bool: every read and write of this field happens while
	// deadlineMu is held, so the mutex is what makes it safe — an atomic
	// type here would wrongly suggest it's safe to check outside deadlineMu,
	// which is exactly the race the deadlineMu introduction closed.
	closed bool
}

// newConnection creates a connection with initialized semaphores.
//
// Always use this constructor: a zero-value channel blocks forever, so a
// missed initialization site turns into a silent permanent deadlock.
func newConnection(conn net.Conn, codec Codec) *connection {
	return &connection{
		conn:     conn,
		codec:    codec,
		readSem:  make(chan struct{}, 1),
		writeSem: make(chan struct{}, 1),
	}
}

// prepareWrite locks deadlineMu, checks closed, and sets the write deadline
// for ctx, all before any blocking I/O. deadlineMu is released via defer, so
// no future error branch added to this method can leave it held — that is
// the reason this is its own method instead of inline code in Send.
func (c *connection) prepareWrite(ctx context.Context) error {
	c.deadlineMu.Lock()
	defer c.deadlineMu.Unlock()

	if c.closed {
		// Wrapped the same way as the codec-write failure in Send: the
		// message was never handed to codec.Write, so retrying it on a new
		// connection is safe, and callers classifying on ErrWriteFailed must
		// see that. net.ErrClosed stays in the chain too.
		return errors.Wrap(writeFailed(net.ErrClosed), "write")
	}
	if err := c.conn.SetWriteDeadline(time.Time{}); err != nil {
		return errors.Wrap(err, "reset write deadline")
	}
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.conn.SetWriteDeadline(deadline); err != nil {
			return errors.Wrap(err, "set write deadline")
		}
	}
	return nil
}

// prepareRead is the Recv counterpart of prepareWrite.
func (c *connection) prepareRead(ctx context.Context) error {
	c.deadlineMu.Lock()
	defer c.deadlineMu.Unlock()

	if c.closed {
		// Unlike prepareWrite's closed path, this must NOT match
		// ErrWriteFailed: a failed Recv is not a write and callers must not
		// treat it as "safe to retry the send".
		return errors.Wrap(net.ErrClosed, "read")
	}
	if err := c.conn.SetReadDeadline(time.Time{}); err != nil {
		return errors.Wrap(err, "reset read deadline")
	}
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.conn.SetReadDeadline(deadline); err != nil {
			return errors.Wrap(err, "set read deadline")
		}
	}
	return nil
}

// Send sends message from buffer using MTProto connection.
func (c *connection) Send(ctx context.Context, b *bin.Buffer) error {
	if err := ctx.Err(); err != nil {
		// Explicit pre-check: with a free semaphore and an already-cancelled
		// ctx, select below would pick a ready case at random, sending on
		// the wire roughly half the time instead of honoring cancellation.
		return errors.Wrap(err, "acquire write")
	}

	// Serializing access to deadlines.
	select {
	case c.writeSem <- struct{}{}:
		defer func() { <-c.writeSem }()
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "acquire write")
	}

	if err := c.prepareWrite(ctx); err != nil {
		return err
	}

	if err := c.codec.Write(c.conn, b); err != nil {
		return errors.Wrap(writeFailed(err), "write")
	}

	return nil
}

// Recv reads message to buffer using MTProto connection.
func (c *connection) Recv(ctx context.Context, b *bin.Buffer) error {
	if err := ctx.Err(); err != nil {
		// See the identical pre-check in Send for why this is needed.
		return errors.Wrap(err, "acquire read")
	}

	// Serializing access to deadlines.
	select {
	case c.readSem <- struct{}{}:
		defer func() { <-c.readSem }()
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "acquire read")
	}

	if err := c.prepareRead(ctx); err != nil {
		return err
	}

	if err := c.codec.Read(c.conn, b); err != nil {
		return errors.Wrap(err, "read")
	}

	return nil
}

// Close closes MTProto connection.
//
// Deadlines are forced into the past before closing: net.Conn.Close is
// documented to unblock pending I/O, but gotd layers transport over arbitrary
// net.Conn implementations (websocket, mtproxy obfuscator) where that is not
// guaranteed. Frame corruption caused by interrupting a write mid-flight is
// irrelevant here — the connection is being discarded.
func (c *connection) Close() error {
	c.deadlineMu.Lock()
	c.closed = true
	_ = c.conn.SetReadDeadline(forcedDeadline)
	_ = c.conn.SetWriteDeadline(forcedDeadline)
	c.deadlineMu.Unlock()
	return c.conn.Close()
}
