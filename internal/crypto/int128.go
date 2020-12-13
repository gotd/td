package crypto

import (
	"io"

	"github.com/gotd/td/bin"
)

// RandInt128 generates and returns new random 128-bit integer.
//
// Use crypto/rand.Reader as randSource in production.
func RandInt128(randSource io.Reader) (bin.Int128, error) {
	buf := make([]byte, bin.Word*4)
	if _, err := io.ReadFull(randSource, buf); err != nil {
		return bin.Int128{}, err
	}
	b := &bin.Buffer{Buf: buf}
	return b.Int128()
}
