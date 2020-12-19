package codec

import (
	"fmt"
	"io"

	"github.com/gotd/td/bin"
)

func tryReadLength(r io.Reader, b *bin.Buffer) (int, error) {
	b.ResetN(bin.Word)
	if _, err := io.ReadFull(r, b.Buf[:bin.Word]); err != nil {
		return 0, fmt.Errorf("failed to read length: %w", err)
	}
	n, err := b.Int()
	if err != nil {
		return 0, err
	}

	if n <= 0 || n > maxMessageSize {
		return 0, errInvalidMsgLen{n: n}
	}

	return n, nil
}
