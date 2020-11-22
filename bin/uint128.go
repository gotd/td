package bin

// Uint128 represents unsigned 128-bit integer.
type Uint128 [2]uint64

func (i *Uint128) Decode(buf *Buffer) error {
	v, err := buf.Uint128()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

func (i Uint128) Encode(b *Buffer) error {
	b.PutUint128(i)
	return nil
}
