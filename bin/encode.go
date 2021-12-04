package bin

import (
	"encoding/binary"
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

// PutUint16 serializes unsigned 16-bit integer.
func (b *Buffer) PutUint16(v uint16) {
	t := make([]byte, 2)
	binary.LittleEndian.PutUint16(t, v)
	b.Buf = append(b.Buf, t...)
}

// PutInt32 serializes signed 32-bit integer.
func (b *Buffer) PutInt32(v int32) {
	b.PutUint32(uint32(v))
}

// PutUint32 serializes unsigned 32-bit integer.
func (b *Buffer) PutUint32(v uint32) {
	t := make([]byte, Word)
	binary.LittleEndian.PutUint32(t, v)
	b.Buf = append(b.Buf, t...)
}

// PutInt53 serializes v as signed integer.
func (b *Buffer) PutInt53(v int64) {
	b.PutLong(v)
}

// PutLong serializes v as signed integer.
func (b *Buffer) PutLong(v int64) {
	b.PutUint64(uint64(v))
}

// PutUint64 serializes v as unsigned 64-bit integer.
func (b *Buffer) PutUint64(v uint64) {
	t := make([]byte, Word*2)
	binary.LittleEndian.PutUint64(t, v)
	b.Buf = append(b.Buf, t...)
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
