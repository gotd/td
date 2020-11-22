package bin

import (
	"errors"
	"math/big"
	"math/bits"
)

// Int128 represents signed 128-bit integer.
type Int128 [2]int64

func (i *Int128) Decode(buf *Buffer) error {
	v, err := buf.Int128()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

func (i Int128) Encode(b *Buffer) error {
	b.PutInt128(i)
	return nil
}

// BigInt returns corresponding big.Int value.
func (i Int128) BigInt() *big.Int {
	v := new(big.Int)
	i.FillBigInt(v)
	return v
}

// FillBigInt sets v to i value.
func (i Int128) FillBigInt(v *big.Int) {
	switch bits.UintSize {
	case 64:
		v.SetBits([]big.Word{
			big.Word(i[0]), // [0; 64)
			big.Word(i[1]), // [64; 128)
		})
	case 32:
		v.SetBits([]big.Word{
			big.Word(i[0]),       // [0; 32)
			big.Word(i[0] >> 32), // [32; 64)
			big.Word(i[1]),       // [64; 96)
			big.Word(i[1] >> 32), // [96; 128)
		})
	default:
		panic("unknown bit size")
	}
}

// ErrIntegerTooBig means that integer provided to SetToBigInt were too big
// to be represented.
var ErrIntegerTooBig = errors.New("td/bin: integer is too big")

func (i *Int128) setToBigInt32(b []big.Word) error {
	if len(b) > 4 {
		return ErrIntegerTooBig
	}
	buf := make([]int32, 4)
	for j := 0; j < len(b); j++ {
		buf[4-len(b)+j] = int32(b[j])
	}

	i[0] = int64(buf[0])
	i[0] |= int64(buf[1]) << 32
	i[1] = int64(buf[2])
	i[1] |= int64(buf[3]) << 32

	return nil
}

func (i *Int128) SetToBigInt(v *big.Int) error {
	b := v.Bits()
	v.Bytes()
	// The big.Word size is platform-dependant.
	switch bits.UintSize {
	case 64:
		if len(b) > 2 {
			return ErrIntegerTooBig
		}
		for j := 0; j < len(b); j++ {
			i[j] = int64(b[0])
		}
		return nil
	case 32:
		return i.setToBigInt32(b)
	default:
		panic("unknown bit size")
	}
}
