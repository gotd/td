package bin

import (
	"encoding/binary"
	"io"
	"math"
)

// PeekID returns next type id in Buffer, but does not consume it.
func (b *Buffer) PeekID() (uint32, error) {
	if len(b.Buf) < Word {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint32(b.Buf)
	return v, nil
}

// PeekN returns n bytes from Buffer to target, but does not consume it.
//
// Returns io.ErrUnexpectedEOF if buffer contains less that n bytes.
// Expects that len(target) >= n.
func (b *Buffer) PeekN(target []byte, n int) error {
	if len(b.Buf) < n {
		return io.ErrUnexpectedEOF
	}
	copy(target, b.Buf[:n])
	return nil
}

// ID decodes type id from Buffer.
func (b *Buffer) ID() (uint32, error) {
	return b.Uint32()
}

// Uint32 decodes unsigned 32-bit integer from Buffer.
func (b *Buffer) Uint32() (uint32, error) {
	v, err := b.PeekID()
	if err != nil {
		return 0, err
	}
	b.Buf = b.Buf[Word:]
	return v, nil
}

// Uint64 decodes 64-bit unsigned integer from Buffer.
func (b *Buffer) Uint64() (uint64, error) {
	const size = Word * 2
	if len(b.Buf) < size {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint64(b.Buf)
	b.Buf = b.Buf[size:]
	return v, nil
}

// Int32 decodes signed 32-bit integer from Buffer.
func (b *Buffer) Int32() (int32, error) {
	v, err := b.Uint32()
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

// ConsumeN consumes n bytes from buffer, writing them to target.
//
// Returns io.ErrUnexpectedEOF if buffer contains less that n bytes.
// Expects that len(target) >= n.
func (b *Buffer) ConsumeN(target []byte, n int) error {
	if err := b.PeekN(target, n); err != nil {
		return err
	}
	b.Buf = b.Buf[n:]
	return nil
}

// Bool decodes bare boolean from Buffer.
func (b *Buffer) Bool() (bool, error) {
	v, err := b.PeekID()
	if err != nil {
		return false, err
	}
	switch v {
	case TypeTrue:
		b.Buf = b.Buf[Word:]
		return true, nil
	case TypeFalse:
		b.Buf = b.Buf[Word:]
		return false, nil
	default:
		return false, NewUnexpectedID(v)
	}
}

// ConsumeID decodes type id from Buffer. If id differs from provided,
// the *UnexpectedIDErr{ID: gotID} will be returned and buffer will be
// not consumed.
func (b *Buffer) ConsumeID(id uint32) error {
	v, err := b.PeekID()
	if err != nil {
		return err
	}
	if v != id {
		return NewUnexpectedID(v)
	}
	b.Buf = b.Buf[Word:]
	return nil
}

// VectorHeader decodes vector length from Buffer.
func (b *Buffer) VectorHeader() (int, error) {
	if err := b.ConsumeID(TypeVector); err != nil {
		return 0, err
	}
	n, err := b.Int()
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, &InvalidLengthError{
			Length: n,
			Where:  "vector",
		}
	}
	return n, nil
}

// String decodes string from Buffer.
func (b *Buffer) String() (string, error) {
	n, v, err := decodeString(b.Buf)
	if err != nil {
		return "", err
	}
	if len(b.Buf) < n {
		return "", io.ErrUnexpectedEOF
	}
	b.Buf = b.Buf[n:]
	return v, nil
}

// Bytes decodes byte slice from Buffer.
//
// NB: returning value is a copy, it's safe to modify it.
func (b *Buffer) Bytes() ([]byte, error) {
	n, v, err := decodeBytes(b.Buf)
	if err != nil {
		return nil, err
	}
	if len(b.Buf) < n {
		return nil, io.ErrUnexpectedEOF
	}
	b.Buf = b.Buf[n:]
	return append([]byte(nil), v...), nil
}

// Int decodes integer from Buffer.
func (b *Buffer) Int() (int, error) {
	v, err := b.Int32()
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// Double decodes 64-bit floating point from Buffer.
func (b *Buffer) Double() (float64, error) {
	v, err := b.Long()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(uint64(v)), nil
}

// Int53 decodes 53-bit signed integer from Buffer.
func (b *Buffer) Int53() (int64, error) {
	return b.Long()
}

// Long decodes 64-bit signed integer from Buffer.
func (b *Buffer) Long() (int64, error) {
	v, err := b.Uint64()
	if err != nil {
		return 0, err
	}
	return int64(v), nil
}

// Int128 decodes 128-bit signed integer from Buffer.
func (b *Buffer) Int128() (Int128, error) {
	v := Int128{}
	size := len(v)
	if len(b.Buf) < size {
		return Int128{}, io.ErrUnexpectedEOF
	}
	copy(v[:size], b.Buf[:size])
	b.Buf = b.Buf[size:]
	return v, nil
}

// Int256 decodes 256-bit signed integer from Buffer.
func (b *Buffer) Int256() (Int256, error) {
	v := Int256{}
	size := len(v)
	if len(b.Buf) < size {
		return Int256{}, io.ErrUnexpectedEOF
	}
	copy(v[:size], b.Buf[:size])
	b.Buf = b.Buf[size:]
	return v, nil
}
