package bin

import (
	"encoding/binary"
	"io"
)

type Decoder struct {
	buf []byte
}

func (d *Decoder) PeekID() (uint32, error) {
	if len(d.buf) < word {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint32(d.buf)
	return v, nil
}

func (d *Decoder) ID() (uint32, error) {
	return d.Uint32()
}

const word = 4

func (d *Decoder) Uint32() (uint32, error) {
	v, err := d.PeekID()
	if err != nil {
		return 0, err
	}
	d.buf = d.buf[word:]
	return v, nil
}
