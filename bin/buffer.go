package bin

import (
	"io"
)

// Buffer implements low level binary (de-)serialization for TL.
type Buffer struct {
	Buf []byte
}

// Encode wrapper.
func (b *Buffer) Encode(e Encoder) error {
	return e.Encode(b)
}

// Decode wrapper.
func (b *Buffer) Decode(d Decoder) error {
	return d.Decode(b)
}

// ResetN resets buffer and expands it to fit n bytes.
func (b *Buffer) ResetN(n int) {
	b.Buf = append(b.Buf[:0], make([]byte, n)...)
}

// Expand expands buffer to add n bytes.
func (b *Buffer) Expand(n int) {
	b.Buf = append(b.Buf, make([]byte, n)...)
}

// Skip moves cursor for next n bytes.
func (b *Buffer) Skip(n int) {
	b.Buf = b.Buf[n:]
}

// Read implements io.Reader.
func (b *Buffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if len(b.Buf) == 0 {
		return 0, io.EOF
	}
	n = copy(p, b.Buf)
	b.Buf = b.Buf[n:]
	return n, nil
}

// Copy returns new copy of buffer.
func (b *Buffer) Copy() []byte {
	return append([]byte{}, b.Buf...)
}

// Raw returns internal byte slice.
func (b Buffer) Raw() []byte {
	return b.Buf
}

// Len returns length of internal buffer.
func (b Buffer) Len() int {
	return len(b.Buf)
}

// ResetTo sets internal buffer exactly to provided value.
//
// Buffer will retain buf, so user should not modify or read it
// concurrently.
func (b *Buffer) ResetTo(buf []byte) {
	b.Buf = buf
}
