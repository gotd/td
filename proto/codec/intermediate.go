package codec

import (
	"io"

	"github.com/go-faster/errors"

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

var (
	_ TaggedCodec = Intermediate{}
)

// WriteHeader sends protocol tag.
func (i Intermediate) WriteHeader(w io.Writer) (err error) {
	if _, err := w.Write(IntermediateClientStart[:]); err != nil {
		return errors.Wrap(err, "write intermediate header")
	}

	return nil
}

// ReadHeader reads protocol tag.
func (i Intermediate) ReadHeader(r io.Reader) (err error) {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return errors.Wrap(err, "read intermediate header")
	}

	if b != IntermediateClientStart {
		return ErrProtocolHeaderMismatch
	}

	return nil
}

// ObfuscatedTag returns protocol tag for obfuscation.
func (i Intermediate) ObfuscatedTag() [4]byte {
	return IntermediateClientStart
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
		return errors.Wrap(err, "write intermediate")
	}

	return nil
}

// Read fills buffer with received message.
func (i Intermediate) Read(r io.Reader, b *bin.Buffer) error {
	if err := readIntermediate(r, b, false); err != nil {
		return errors.Wrap(err, "read intermediate")
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
		return errors.Wrap(err, "read payload")
	}

	if padding {
		paddingLength := n % 4
		b.Buf = b.Buf[:n-paddingLength]
	}

	return nil
}
