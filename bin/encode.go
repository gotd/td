package bin

import (
	"math"
)

// PutID serializes type definition id, like a8509bda.
func (b *Buffer) PutID(id uint32) {
	b.PutUint32(id)
}

// Put appends raw bytes to buffer.
//
// Buffer does not retain raw.
func (b *Buffer) Put(raw []byte) {
	b.Buf = append(b.Buf, raw...)
}

// PutPadding appends zeroes to buffer as padding.
func (b *Buffer) PutPadding(n int) {
	b.Buf = append(b.Buf, make([]byte, n)...)
}

// PutString serializes bare string.
func (b *Buffer) PutString(s string) {
	b.Buf = encodeString(b.Buf, s)
}

// PutBytes serializes bare byte string.
func (b *Buffer) PutBytes(v []byte) {
	b.Buf = encodeBytes(b.Buf, v)
}

// PutVectorHeader serializes vector header with provided length.
func (b *Buffer) PutVectorHeader(length int) {
	b.PutID(TypeVector)
	b.PutInt32(int32(length))
}

// PutInt serializes v as signed 32-bit integer.
//
// If v is bigger than 32-bit, `behavior` is undefined.
func (b *Buffer) PutInt(v int) {
	b.PutInt32(int32(v))
}

// PutBool serializes bare boolean.
func (b *Buffer) PutBool(v bool) {
	switch v {
	case true:
		b.PutID(TypeTrue)
	case false:
		b.PutID(TypeFalse)
	}
}

// PutInt32 serializes signed 32-bit integer.
func (b *Buffer) PutInt32(v int32) {
	b.Buf = append(b.Buf,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
	)
}

func (b *Buffer) PutUint32(v uint32) {
	b.Buf = append(b.Buf,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
	)
}

// PutLong serializes v as signed integer.
func (b *Buffer) PutLong(v int64) {
	b.Buf = append(b.Buf,
		byte(v),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
		byte(v>>32),
		byte(v>>40),
		byte(v>>48),
		byte(v>>56),
	)
}

// PutUint64 serializes v as unsigned 64-bit integer.
func (b *Buffer) PutUint64(v uint64) {
	b.Buf = append(b.Buf,
		byte(v),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
		byte(v>>32),
		byte(v>>40),
		byte(v>>48),
		byte(v>>56),
	)
}

// PutDouble serializes v as 64-bit floating point.
func (b *Buffer) PutDouble(v float64) {
	b.PutUint64(math.Float64bits(v))
}

// PutInt128 serializes v as 128-bit signed integer.
func (b *Buffer) PutInt128(v Int128) {
	b.Buf = append(b.Buf, v[:]...)
}

// PutInt256 serializes v as 256-bit signed integer.
func (b *Buffer) PutInt256(v Int256) {
	b.Buf = append(b.Buf, v[:]...)
}

// Reset buffer to zero length.
func (b *Buffer) Reset() {
	b.Buf = b.Buf[:0]
}
