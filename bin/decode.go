package bin

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// PeekID returns next type id in Buffer, but does not consume it.
func (b *Buffer) PeekID() (uint32, error) {
	if len(b.buf) < Word {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint32(b.buf)
	return v, nil
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
	b.buf = b.buf[Word:]
	return v, nil
}

// Int32 decodes signed 32-bit integer from Buffer.
func (b *Buffer) Int32() (int32, error) {
	if len(b.buf) < Word {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint32(b.buf)
	b.buf = b.buf[Word:]
	return int32(v), nil
}

// UnexpectedIDErr means that unknown or unexpected type id was decoded.
type UnexpectedIDErr struct {
	ID uint32
}

func (e UnexpectedIDErr) Error() string {
	return fmt.Sprintf("unexpected id 0x%x", e.ID)
}

// NewUnexpectedID return new UnexpectedIDErr.
func NewUnexpectedID(id uint32) error {
	return &UnexpectedIDErr{ID: id}
}

// Bool decodes bare boolean from Buffer.
func (b *Buffer) Bool() (bool, error) {
	v, err := b.PeekID()
	if err != nil {
		return false, err
	}
	switch v {
	case TypeTrue:
		b.buf = b.buf[Word:]
		return true, nil
	case TypeFalse:
		b.buf = b.buf[Word:]
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
	b.buf = b.buf[Word:]
	return nil
}

// VectorHeader decodes vector length from Buffer.
func (b *Buffer) VectorHeader() (int, error) {
	id, err := b.PeekID()
	if err != nil {
		return 0, err
	}
	if id != TypeVector {
		return 0, NewUnexpectedID(id)
	}
	b.buf = b.buf[Word:]
	n, err := b.Int32()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// String decodes string from Buffer.
func (b *Buffer) String() (string, error) {
	n, v, err := decodeString(b.buf)
	if err != nil {
		return "", err
	}
	b.buf = b.buf[n:]
	return v, nil
}

// Bytes decodes byte slice from Buffer.
//
// NB: returning value is slice of buffer, it is not safe
// to retain or modify. User should copy value if needed.
func (b *Buffer) Bytes() ([]byte, error) {
	n, v, err := decodeBytes(b.buf)
	if err != nil {
		return nil, err
	}
	b.buf = b.buf[n:]
	return v, nil
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

// Long decodes 64-bit signed integer from Buffer.
func (b *Buffer) Long() (int64, error) {
	const size = Word * 2
	if len(b.buf) < size {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint64(b.buf)
	b.buf = b.buf[size:]
	return int64(v), nil
}
