package crypto

import (
	"io"

	"github.com/gotd/td/bin"
)

// RandInt256 generates and returns new random 256-bit integer.
//
// Use crypto/rand.Reader as randSource in production.
func RandInt256(randSource io.Reader) (bin.Int256, error) {
	buf := make([]byte, bin.Word*8)
	if _, err := io.ReadFull(randSource, buf); err != nil {
		return bin.Int256{}, err
	}
	b := &bin.Buffer{Buf: buf}
	return b.Int256()
}
