package bin

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytesDecodeEncode(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		a := require.New(t)

		for _, b := range [][]byte{
			bytes.Repeat([]byte{1, 2, 3, 4}, 100),
			bytes.Repeat([]byte{1, 2, 3}, 102),
			bytes.Repeat([]byte{1, 2}, 103),
			bytes.Repeat([]byte{10}, 104),
			bytes.Repeat([]byte{6}, 105),
			[]byte("foo"),
			[]byte("b"),
			[]byte("ba"),
			[]byte("what are you doing?"),
			[]byte("кек"),
			{
				0x57, 0x61, 0x6b, 0x65,
				0x20, 0x75, 0x70, 0x2c,
				0x20, 0x4e, 0x65, 0x6f,
			},
			bytes.Repeat([]byte{1}, 253),
		} {
			buf := encodeBytes(nil, b)
			a.True(len(buf)%4 != 0, "bad align")

			n, v, err := decodeBytes(buf)
			a.NoError(err)
			a.Equal(b, v)
			a.NotZero(n, "zero bytes read return")
		}
	})
	t.Run("ErrUnexpectedEOF", func(t *testing.T) {
		a := require.New(t)

		_, _, err := decodeBytes(nil)
		a.ErrorIs(err, io.ErrUnexpectedEOF)

		_, _, err = decodeBytes([]byte{firstLongStringByte})
		a.ErrorIs(err, io.ErrUnexpectedEOF)

		_, _, err = decodeBytes([]byte{firstLongStringByte, 0, 0, 255, 0})
		a.ErrorIs(err, io.ErrUnexpectedEOF)

		_, _, err = decodeBytes(encodeString(nil, "foo bar")[:2])
		a.ErrorIs(err, io.ErrUnexpectedEOF)

		_, _, err = decodeBytes(encodeString(nil, strings.Repeat("b", 105))[:10])
		a.ErrorIs(err, io.ErrUnexpectedEOF)
	})
	t.Run("InvalidLength", func(t *testing.T) {
		_, _, err := decodeBytes(bytes.Repeat([]byte{255}, 256))
		require.ErrorIs(t, err, errInvalidLength)
	})
}
