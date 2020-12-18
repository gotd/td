package proto

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
)

// The Intermediate version of MTproto.
//
// See https://core.telegram.org/mtproto/mtproto-transports#intermediate

// IntermediateClientStart is starting bytes sent by client in Intermediate mode.
//
// Note that server does not respond with it.
var IntermediateClientStart = []byte{0xee, 0xee, 0xee, 0xee}

// Intermediate is intermediate MTProto transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#intermediate
type Intermediate struct {
	Dialer Dialer

	conn net.Conn
}

// check that Intermediate implements Transport in compile time.
var _ Transport = &Intermediate{}

// IntermediateFromConnection creates Intermediate transport fron given net.Conn
// For tgtest.Server purposes only.
func IntermediateFromConnection(conn net.Conn) *Intermediate {
	return &Intermediate{conn: conn}
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

	if err := writeIntermediate(i.conn, b); err != nil {
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

	if err := readIntermediate(i.conn, b); err != nil {
		return xerrors.Errorf("read intermediate: %w", err)
	}

	if err := i.conn.SetReadDeadline(time.Time{}); err != nil {
		return xerrors.Errorf("reset connection deadline: %w", err)
	}

	if err := checkProtocolError(b); err != nil {
		return err
	}

	return nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (i *Intermediate) Close() error {
	return i.conn.Close()
}

type errInvalidMsgLen struct {
	n int
}

func (e errInvalidMsgLen) Error() string {
	return fmt.Sprintf("invalid message length %d", e.n)
}

func (e errInvalidMsgLen) Is(err error) bool {
	_, ok := err.(errInvalidMsgLen)
	return ok
}

// writeIntermediate encodes b as payload to w.
func writeIntermediate(w io.Writer, b *bin.Buffer) error {
	if b.Len() > maxMessageSize {
		return errInvalidMsgLen{n: b.Len()}
	}

	// Re-using b.Buf if possible to reduce allocations.
	buf := append(b.Buf[len(b.Buf):], make([]byte, 0, 4)...)
	inner := bin.Buffer{Buf: buf}
	inner.PutInt(b.Len())
	if _, err := w.Write(inner.Buf); err != nil {
		return err
	}
	if _, err := w.Write(b.Raw()); err != nil {
		return err
	}
	return nil
}

const maxMessageSize = 1024 * 1024 // 1mb

// readIntermediate reads payload from r to b.
func readIntermediate(r io.Reader, b *bin.Buffer) error {
	b.ResetN(bin.Word)
	if _, err := io.ReadFull(r, b.Buf); err != nil {
		return fmt.Errorf("failed to read length: %w", err)
	}
	n, err := b.Int()
	if err != nil {
		return err
	}

	if n <= 0 || n > maxMessageSize {
		return errInvalidMsgLen{n: n}
	}
	b.ResetN(n)
	if _, err := io.ReadFull(r, b.Buf); err != nil {
		return fmt.Errorf("failed to read payload: %w", err)
	}

	return nil
}
