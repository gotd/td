package proto

import (
	"errors"
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

// WriteIntermediate encodes b as payload to w.
func WriteIntermediate(w io.Writer, b *bin.Buffer) error {
	if b.Len() > maxMessageSize {
		return ErrMessageTooBig
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

// ErrMessageTooBig means that message length is too big to be handled.
var ErrMessageTooBig = errors.New("message is too big")

const maxMessageSize = 1024 * 1024 // 1mb

// ErrMessageTooSmall means that message length is too small and invalid.
var ErrMessageTooSmall = errors.New("message is too small")

const minMessageSize = 0

// ReadIntermediate reads payload from r to b.
func ReadIntermediate(r io.Reader, b *bin.Buffer) error {
	b.PreAllocate(4)
	if _, err := io.ReadFull(r, b.Buf[:4]); err != nil {
		return fmt.Errorf("failed to read length: %w", err)
	}
	dataLen, err := b.Int32()
	if err != nil {
		return err
	}

	switch {
	case dataLen < minMessageSize:
		return ErrMessageTooSmall
	case dataLen > maxMessageSize:
		return ErrMessageTooBig
	}

	b.PreAllocate(int(dataLen))
	if _, err := io.ReadFull(r, b.Buf[0:int(dataLen)]); err != nil {
		return fmt.Errorf("failed to read payload: %w", err)
	}

	return nil
}
