package codec

import (
	"encoding/binary"
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

// Codec is MTProto transport protocol encoding abstraction.
type Codec interface {
	// WriteHeader sends protocol tag if needed.
	WriteHeader(w io.Writer) error
	// ReadHeader reads protocol tag if needed.
	ReadHeader(r io.Reader) error
	// Write encode to writer message from given buffer.
	Write(w io.Writer, b *bin.Buffer) error
	// Read fills buffer with received message.
	Read(r io.Reader, b *bin.Buffer) error
}

// TaggedCodec is codec with protocol tag.
type TaggedCodec interface {
	Codec
	// ObfuscatedTag returns protocol tag for obfuscation.
	ObfuscatedTag() [4]byte
}

// readLen reads 32-bit integer and validates it as message length.
func readLen(r io.Reader, b *bin.Buffer) (int, error) {
	b.ResetN(bin.Word)
	if _, err := io.ReadFull(r, b.Buf[:bin.Word]); err != nil {
		return 0, errors.Wrap(err, "read length")
	}
	n := int(binary.LittleEndian.Uint32(b.Buf[:bin.Word]))

	if n <= 0 || n > maxMessageSize {
		return 0, invalidMsgLenErr{n: n}
	}

	return n, nil
}
