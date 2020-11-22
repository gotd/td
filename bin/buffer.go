package bin

// Buffer implements binary serialization for TL.
type Buffer struct {
	buf []byte
}

// Raw returns internal byte slice.
func (b Buffer) Raw() []byte {
	return b.buf
}

func (b *Buffer) ResetTo(buf []byte) {
	b.buf = buf
}
