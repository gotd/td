package crypto

import (
	"io"

	"github.com/ernado/td/bin"
)

func RandInt128(reader io.Reader) (bin.Int128, error) {
	buf := make([]byte, bin.Word*4)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return bin.Int128{}, err
	}
	b := &bin.Buffer{Buf: buf}
	return b.Int128()
}
