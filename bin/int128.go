package bin

import "math/big"

// Int128 represents signed 128-bit integer.
type Int128 [16]byte

// Decode implements bin.Decoder.
func (i *Int128) Decode(buf *Buffer) error {
	v, err := buf.Int128()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

// Encode implements bin.Encoder.
func (i Int128) Encode(b *Buffer) error {
	b.PutInt128(i)
	return nil
}

// BigInt returns corresponding big.Int value.
func (i Int128) BigInt() *big.Int {
	return big.NewInt(0).SetBytes(i[:])
}
