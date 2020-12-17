package proto

import (
	"fmt"
	"io"

	"github.com/gotd/td/bin"
)

// The Intermediate version of MTproto.
//
// See https://core.telegram.org/mtproto/mtproto-transports#intermediate

// IntermediateClientStart is starting bytes sent by client in Intermediate mode.
//
// Note that server does not respond with it.
var IntermediateClientStart = []byte{0xee, 0xee, 0xee, 0xee}

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

// WriteIntermediate encodes b as payload to w.
func WriteIntermediate(w io.Writer, b *bin.Buffer) error {
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

// ReadIntermediate reads payload from r to b.
func ReadIntermediate(r io.Reader, b *bin.Buffer) error {
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
