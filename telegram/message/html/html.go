// Package html contains HTML styling options.
package html

import (
	"bytes"
	"io"
	"strings"

	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
)

// Bytes reads HTML from given byte slice and returns styling option
// to build styled text block.
func Bytes(b []byte) styling.StyledTextOption {
	return Reader(bytes.NewReader(b))
}

// String reads HTML from given string and returns styling option
// to build styled text block.
func String(s string) styling.StyledTextOption {
	return Reader(strings.NewReader(s))
}

// Reader reads HTML from given reader and returns styling option
// to build styled text block.
func Reader(r io.Reader) styling.StyledTextOption {
	return styling.Custom(func(eb *entity.Builder) error {
		return entity.HTML(r, eb)
	})
}
