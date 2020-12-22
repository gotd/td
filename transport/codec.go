package transport

import (
	"io"

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
