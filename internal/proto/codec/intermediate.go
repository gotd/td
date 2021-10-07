package codec

import (
	"io"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
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

// ObfuscatedTag returns protocol tag for obfuscation.
func (i Intermediate) ObfuscatedTag() (r []byte) {
	return append(r, IntermediateClientStart[:]...)
}

// Write encode to writer message from given buffer.
func (i Intermediate) Write(w io.Writer, b *bin.Buffer) error {
	if err := checkOutgoingMessage(b); err != nil {
		return err
	}

	if err := checkAlign(b, 4); err != nil {
		return err
	}

	if err := writeIntermediate(w, b); err != nil {
		return xerrors.Errorf("write intermediate: %w", err)
	}

	return nil
}

// Read fills buffer with received message.
func (i Intermediate) Read(r io.Reader, b *bin.Buffer) error {
	if err := readIntermediate(r, b, false); err != nil {
		return xerrors.Errorf("read intermediate: %w", err)
	}

	return checkProtocolError(b)
}

// writeIntermediate encodes b as payload to w.
func writeIntermediate(w io.Writer, b *bin.Buffer) error {
	length := b.Len()
	// Re-using b.Buf if possible to reduce allocations.
	b.Expand(4)
	b.Buf = b.Buf[:length]

	inner := bin.Buffer{Buf: b.Buf[length:length]}
	inner.PutInt(b.Len())
	if _, err := w.Write(inner.Buf); err != nil {
		return err
	}
	if _, err := w.Write(b.Buf); err != nil {
		return err
	}
	return nil
}

// readIntermediate reads payload from r to b.
func readIntermediate(r io.Reader, b *bin.Buffer, padding bool) error {
	n, err := readLen(r, b)
	if err != nil {
		return err
	}

	b.ResetN(n)
	if _, err := io.ReadFull(r, b.Buf); err != nil {
		return xerrors.Errorf("read payload: %w", err)
	}

	if padding {
		paddingLength := n % 4
		b.Buf = b.Buf[:n-paddingLength]
	}

	return nil
}
