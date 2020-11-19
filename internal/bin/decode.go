package bin

import (
	"encoding/binary"
	"fmt"
	"io"
)

func (b *Buffer) PeekID() (uint32, error) {
	if len(b.buf) < word {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint32(b.buf)
	return v, nil
}

func (b *Buffer) ID() (uint32, error) {
	return b.Uint32()
}

const word = 4

func (b *Buffer) Uint32() (uint32, error) {
	v, err := b.PeekID()
	if err != nil {
		return 0, err
	}
	b.buf = b.buf[word:]
	return v, nil
}

func (b *Buffer) Int32() (int32, error) {
	if len(b.buf) < word {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint32(b.buf)
	b.buf = b.buf[word:]
	return int32(v), nil
}

type UnexpectedIDErr struct {
	ID uint32
}

func (e UnexpectedIDErr) Error() string {
	return fmt.Sprintf("unexpected id 0x%x", e.ID)
}

func NewUnexpectedID(id uint32) error {
	return &UnexpectedIDErr{ID: id}
}

func (b *Buffer) Bool() (bool, error) {
	v, err := b.PeekID()
	if err != nil {
		return false, err
	}
	switch v {
	case TypeTrue:
		b.buf = b.buf[word:]
		return true, nil
	case TypeFalse:
		b.buf = b.buf[word:]
		return false, nil
	default:
		return false, NewUnexpectedID(v)
	}
}

func (b *Buffer) ConsumeID(id uint32) error {
	v, err := b.PeekID()
	if err != nil {
		return err
	}
	if v != id {
		return NewUnexpectedID(v)
	}
	b.buf = b.buf[word:]
	return nil
}

func (b *Buffer) VectorHeader() (int, error) {
	id, err := b.PeekID()
	if err != nil {
		return 0, err
	}
	if id != TypeVector {
		return 0, NewUnexpectedID(id)
	}
	b.buf = b.buf[word:]
	n, err := b.Int32()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (b *Buffer) String() (string, error) {
	n, v, err := decodeString(b.buf)
	if err != nil {
		return "", err
	}
	b.buf = b.buf[n:]
	return v, nil
}
