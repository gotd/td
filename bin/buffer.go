package bin

// Buffer implements low level binary (de-)serialization for TL.
type Buffer struct {
	buf []byte
}

// Raw returns internal byte slice.
func (b Buffer) Raw() []byte {
	return b.buf
}

// ResetTo sets internal buffer exactly to provided value.
//
// Buffer will retain buf, so user should not modify or read it
// concurrently.
func (b *Buffer) ResetTo(buf []byte) {
	b.buf = buf
}
