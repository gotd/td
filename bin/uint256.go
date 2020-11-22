package bin

// Uint128 represents unsigned 256-bit integer.
type Uint256 [4]uint64

func (i *Uint256) Decode(buf *Buffer) error {
	v, err := buf.Uint256()
	if err != nil {
		return err
	}
	*i = v
	return nil
}

func (i Uint256) Encode(b *Buffer) error {
	b.PutUint256(i)
	return nil
}
