package crypto

import "math/big"

// FillBytes is safe version of (*big.Int).FillBytes.
// Returns false if to length is not exact equal to big.Int's.
// Otherwise fills to using b and returns true.
func FillBytes(b *big.Int, to []byte) bool {
	bits := b.BitLen()
	if (bits+7)/8 > len(to) {
		return false
	}
	b.FillBytes(to)
	return true
}
