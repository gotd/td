package bin

import (
	"io"
)

// Buffer implements low level binary (de-)serialization for TL.
type Buffer struct {
	Buf []byte
}

// PreAllocate reallocates buffer if it is not too big enough.
// Returns true, if reallocation happen.
func (b *Buffer) PreAllocate(size int) (r bool) {
	if len(b.Buf) < size {
		n := make([]byte, size)
		copy(n, b.Buf)
		b.Buf = n

		r = true
	}

	return
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

// Write implements io.Writer.
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.Buf = append(b.Buf, p...)
	return len(p), nil
}

// WriteTo implements io.WriterTo.
func (b Buffer) WriteTo(w io.Writer) (n int64, err error) {
	wroteN, err := w.Write(b.Buf)
	return int64(wroteN), err
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
