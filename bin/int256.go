package bin

import "math/big"

// Int256 represents signed 256-bit integer.
type Int256 [32]byte

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

func (i Int256) BigInt() *big.Int {
	return big.NewInt(0).SetBytes(i[:])
}
