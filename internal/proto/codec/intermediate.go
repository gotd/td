package codec

import (
	"fmt"
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
)

// IntermediateClientStart is starting bytes sent by client in Intermediate mode.
//
// Note that server does not respond with it.
var IntermediateClientStart = [4]byte{0xee, 0xee, 0xee, 0xee}

// Intermediate is intermediate MTProto transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#intermediate
type Intermediate struct{}

// WriteHeader sends protocol tag.
func (i Intermediate) WriteHeader(w io.Writer) (err error) {
	if _, err := w.Write(IntermediateClientStart[:]); err != nil {
		return xerrors.Errorf("write intermediate header: %w", err)
	}

	return nil
}

// ReadHeader reads protocol tag.
func (i Intermediate) ReadHeader(r io.Reader) (err error) {
	var b [4]byte
	if _, err := r.Read(b[:]); err != nil {
		return xerrors.Errorf("read intermediate header: %w", err)
	}

	if b != IntermediateClientStart {
		return ErrProtocolHeaderMismatch
	}

	return nil
}

// Write encode to writer message from given buffer.
func (i Intermediate) Write(w io.Writer, b *bin.Buffer) error {
	if err := writeIntermediate(w, b); err != nil {
		return xerrors.Errorf("write intermediate: %w", err)
	}

	return nil
}

// Read fills buffer with received message.
func (i Intermediate) Read(r io.Reader, b *bin.Buffer) error {
	if err := readIntermediate(r, b); err != nil {
		return xerrors.Errorf("read intermediate: %w", err)
	}

	if err := checkProtocolError(b); err != nil {
		return err
	}

	return nil
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

// readIntermediate reads payload from r to b.
func readIntermediate(r io.Reader, b *bin.Buffer) error {
	n, err := readLen(r, b)
	if err != nil {
		return err
	}

	b.ResetN(n)
	if _, err := io.ReadFull(r, b.Buf); err != nil {
		return fmt.Errorf("failed to read payload: %w", err)
	}

	return nil
}
