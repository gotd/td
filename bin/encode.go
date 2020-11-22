package bin

import "math"

// PutID serializes type definition id, like a8509bda.
func (b *Buffer) PutID(id uint32) {
	b.PutUint32(id)
}

// Put appends raw bytes to buffer.
//
// Buffer does not retain raw.
func (b *Buffer) Put(raw []byte) {
	b.buf = append(b.buf, raw...)
}

// PutString serializes bare string.
func (b *Buffer) PutString(s string) {
	b.buf = encodeString(b.buf, s)
}

// PutBytes serializes bare byte string.
func (b *Buffer) PutBytes(v []byte) {
	b.buf = encodeBytes(b.buf, v)
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
	b.buf = append(b.buf,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
	)
}

func (b *Buffer) PutUint32(v uint32) {
	b.buf = append(b.buf,
		byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
	)
}

// PutLong serializes v as signed integer.
func (b *Buffer) PutLong(v int64) {
	b.buf = append(b.buf,
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
	b.buf = append(b.buf,
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

// PutUint128 serializes v as 128-bit unsigned integer.
func (b *Buffer) PutUint128(v Uint128) {
	b.buf = append(b.buf,
		byte(v[0]),
		byte(v[0]>>8),
		byte(v[0]>>16),
		byte(v[0]>>24),
		byte(v[0]>>32),
		byte(v[0]>>40),
		byte(v[0]>>48),
		byte(v[0]>>56),

		byte(v[1]),
		byte(v[1]>>8),
		byte(v[1]>>16),
		byte(v[1]>>24),
		byte(v[1]>>32),
		byte(v[1]>>40),
		byte(v[1]>>48),
		byte(v[1]>>56),
	)
}

// PutUint128 serializes v as 256-bit unsigned integer.
func (b *Buffer) PutUint256(v Uint256) {
	b.PutUint128([2]uint64{v[0], v[1]})
	b.PutUint128([2]uint64{v[2], v[3]})
}

// Reset buffer to zero length.
func (b *Buffer) Reset() {
	b.buf = b.buf[:0]
}
