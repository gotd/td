package downloader

import (
	"bytes"
	"image/jpeg"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/testutil"
)

var strippedImage = []uint8{
	0x01, 0x0e, 0x28, 0xa3, 0x9e, 0x05, 0x26, 0x78, 0xa5, 0x03, 0x8e, 0xb4, 0xd2, 0x31, 0x40, 0x06,
	0x7d, 0x85, 0x19, 0xa4, 0xe2, 0x8e, 0x28, 0x00, 0xa2, 0x8a, 0x28, 0x03,
}

func TestExpandThumbnail(t *testing.T) {
	a := require.New(t)

	to := make([]byte, 0, 1024)
	testutil.ZeroAlloc(t, func() {
		var err error
		to, err = ExpandThumbnail(strippedImage, to)
		a.NoError(err)
	})

	_, err := jpeg.Decode(bytes.NewReader(to))
	a.NoError(err)
}
