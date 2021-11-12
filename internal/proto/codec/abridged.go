package codec

import (
	"encoding/binary"
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

// AbridgedClientStart is starting bytes sent by client in Abridged mode.
//
// Note that server does not respond with it.
var AbridgedClientStart = [1]byte{0xef}

// Abridged is intermediate MTProto transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#abridged
type Abridged struct{}

var (
	_ TaggedCodec = Abridged{}
)

// WriteHeader sends protocol tag.
func (i Abridged) WriteHeader(w io.Writer) error {
	if _, err := w.Write(AbridgedClientStart[:]); err != nil {
		return errors.Wrap(err, "write abridged header")
	}

	return nil
}

// ReadHeader reads protocol tag.
func (i Abridged) ReadHeader(r io.Reader) error {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return errors.Wrap(err, "read abridged header")
	}

	if b != AbridgedClientStart {
		return ErrProtocolHeaderMismatch
	}

	return nil
}

// ObfuscatedTag returns protocol tag for obfuscation.
func (i Abridged) ObfuscatedTag() (r [4]byte) {
	d := AbridgedClientStart[0]
	return [4]byte{d, d, d, d}
}

// Write encode to writer message from given buffer.
func (i Abridged) Write(w io.Writer, b *bin.Buffer) error {
	if err := checkOutgoingMessage(b); err != nil {
		return err
	}

	if err := checkAlign(b, 4); err != nil {
		return err
	}

	if err := writeAbridged(w, b); err != nil {
		return errors.Wrap(err, "write abridged")
	}

	return nil
}

// Read fills buffer with received message.
func (i Abridged) Read(r io.Reader, b *bin.Buffer) error {
	if err := readAbridged(r, b); err != nil {
		return errors.Wrap(err, "read abridged")
	}

	return checkProtocolError(b)
}

func writeAbridged(w io.Writer, b *bin.Buffer) error {
	length := b.Len()
	// Re-using b.Buf if possible to reduce allocations.
	b.Expand(4)
	b.Buf = b.Buf[:length]

	// Re-using b.Buf if possible to reduce allocations.
	inner := bin.Buffer{Buf: b.Buf[length:length]}

	encodeLength := b.Len() >> 2
	// `0x7f == 127`, literally use one bit to distinguish length byte size.
	if encodeLength < 127 {
		// Payloads are wrapped in the following envelope:
		//
		// Length: payload length, divided by four, and encoded as a single byte,
		// only if the resulting packet length is a value between 0x01..0x7e.
		inner.Put([]byte{byte(encodeLength)})
	} else {
		// If the packet length divided by four is bigger than or equal to 127 (>= 0x7f),
		// the following envelope must be used, instead:
		//
		var buf [5]byte
		// Header: A single byte of value 0x7f
		buf[0] = 0x7f
		// Length: payload length, divided by four, and encoded as 3 length bytes (little endian)
		binary.LittleEndian.PutUint32(buf[1:], uint32(encodeLength))
		inner.Put(buf[:4])
	}

	if _, err := w.Write(inner.Buf); err != nil {
		return err
	}
	if _, err := w.Write(b.Raw()); err != nil {
		return err
	}
	return nil
}

func readAbridged(r io.Reader, b *bin.Buffer) error {
	b.ResetN(bin.Word)

	_, err := io.ReadFull(r, b.Buf[:1])
	if err != nil {
		return err
	}

	if b.Buf[0] >= 127 {
		_, err := io.ReadFull(r, b.Buf[0:3])
		if err != nil {
			return err
		}
	}

	n, err := b.Int()
	if err != nil {
		return err
	}

	b.ResetN(n << 2)
	if _, err := io.ReadFull(r, b.Buf); err != nil {
		return errors.Wrap(err, "read payload")
	}

	return nil
}
