package crypto

import (
	"io"

	"github.com/gotd/td/bin"
)

// RandInt64 returns random int64 from randSource.
func RandInt64(randSource io.Reader) (int64, error) {
	buf := make([]byte, bin.Word*4)
	if _, err := io.ReadFull(randSource, buf); err != nil {
		return 0, err
	}
	b := &bin.Buffer{Buf: buf}
	return b.Long()
}
