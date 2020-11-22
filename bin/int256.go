package bin

import (
	"math/big"
	"math/bits"
)

// Int256 represents signed 256-bit integer.
type Int256 [4]int64

func (i *Int256) Decode(buf *Buffer) error {
	v, err := buf.Int256()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

func (i Int256) Encode(b *Buffer) error {
	b.PutInt256(i)
	return nil
}

// FillBigInt sets v to i value.
func (i Int256) FillBigInt(v *big.Int) {
	switch bits.UintSize {
	case 64:
		v.SetBits([]big.Word{
			big.Word(i[0]), // [0; 64)
			big.Word(i[1]), // [64; 128)
			big.Word(i[2]), // [64; 128)
			big.Word(i[3]), // [64; 128)
		})
	case 32:
		v.SetBits([]big.Word{
			big.Word(i[0]),       // [0; 32)
			big.Word(i[0] >> 32), // [32; 64)
			big.Word(i[1]),       // [64; 96)
			big.Word(i[1] >> 32), // [96; 128)
			big.Word(i[2]),       // [128; 160)
			big.Word(i[2] >> 32), // [160; 194)
			big.Word(i[3]),       // [194; 224)
			big.Word(i[3] >> 32), // [224; 256)
		})
	default:
		panic("unknown bit size")
	}
}
