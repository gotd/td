package bin

import "math/big"

// Int256 represents signed 256-bit integer.
type Int256 [32]byte

// Decode implements bin.Decoder.
func (i *Int256) Decode(buf *Buffer) error {
	v, err := buf.Int256()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// Encode implements bin.Encoder.
func (i Int256) Encode(b *Buffer) error {
	b.PutInt256(i)
	return nil
}

// BigInt returns corresponding big.Int value.
func (i Int256) BigInt() *big.Int {
	return big.NewInt(0).SetBytes(i[:])
}
