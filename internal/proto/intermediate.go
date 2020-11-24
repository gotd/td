package proto

import (
	"fmt"
	"io"

	"github.com/ernado/td/bin"
)

// The Intermediate version of MTproto.
//
// See https://core.telegram.org/mtproto/mtproto-transports#intermediate

// IntermediateClientStart is starting bytes sent by client in Intermediate mode.
//
// Note that server does not respond with it.
var IntermediateClientStart = []byte{0xee, 0xee, 0xee, 0xee}

// EncodeIntermediate encodes b as payload to w.
func WriteIntermediate(w io.Writer, b *bin.Buffer) error {
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

// ReadIntermediate reads payload from r to b.
func ReadIntermediate(r io.Reader, b *bin.Buffer) error {
	b.Buf = append(b.Buf[:0], make([]byte, 4)...)
	if _, err := io.ReadFull(r, b.Buf[:4]); err != nil {
		return fmt.Errorf("failed to read length: %w", err)
	}
	dataLen, err := b.Int32()
	if err != nil {
		return err
	}
	b.Buf = append(b.Buf[:0], make([]byte, int(dataLen))...)
	if _, err := r.Read(b.Buf); err != nil {
		return fmt.Errorf("failed to read payload: %w", err)
	}
	return nil
}
