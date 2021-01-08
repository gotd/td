package bin

import "strconv"

// Fields represent a bitfield value that compactly encodes
// information about provided conditional fields, e.g. says
// that fields "1", "5" and "10" were set.
type Fields uint32

// Zero returns true, if all bits are equal to zero.
func (f Fields) Zero() bool {
	return f == 0
}

// String implement fmt.Stringer
func (f Fields) String() string {
	return strconv.FormatUint(uint64(f), 2)
}

// Decode implements Decoder.
func (f *Fields) Decode(b *Buffer) error {
	v, err := b.Int32()
	if err != nil {
		return err
	}
	*f = Fields(v)
	return nil
}

// Encode implements Encoder.
func (f Fields) Encode(b *Buffer) error {
	b.PutUint32(uint32(f))
	return nil
}

// Has reports whether field with index n was set.
func (f Fields) Has(n int) bool {
	return f&(1<<n) != 0
}

// Unset unsets field with index n.
func (f *Fields) Unset(n int) {
	*f &= ^(1 << n)
}

// Set sets field with index n.
func (f *Fields) Set(n int) {
	*f |= 1 << n
}
