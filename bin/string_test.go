package bin

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringDecodeEncode(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		a := require.New(t)

		for _, s := range []string{
			strings.Repeat("abcd", 100),
			strings.Repeat("abc", 102),
			strings.Repeat("de", 103),
			strings.Repeat("z", 104),
			strings.Repeat("b", 105),
			"foo",
			"b",
			"ba",
			"what are you doing?",
			"кек",
			strings.Repeat("a", 253),
		} {
			buf := encodeString(nil, s)
			a.True(len(buf)%4 == 0, "bad align")

			n, v, err := decodeString(buf)
			a.NoError(err)
			a.Equal(s, v)
			a.NotZero(n, "zero bytes read return")
		}
	})
	t.Run("ErrUnexpectedEOF", func(t *testing.T) {
		a := require.New(t)

		_, _, err := decodeString(nil)
		a.ErrorIs(err, io.ErrUnexpectedEOF)

		_, _, err = decodeString([]byte{firstLongStringByte})
		a.ErrorIs(err, io.ErrUnexpectedEOF)

		_, _, err = decodeString([]byte{firstLongStringByte, 0, 0, 255, 0})
		a.ErrorIs(err, io.ErrUnexpectedEOF)

		_, _, err = decodeString(encodeString(nil, "foo bar")[:2])
		a.ErrorIs(err, io.ErrUnexpectedEOF)

		_, _, err = decodeString(encodeString(nil, strings.Repeat("b", 105))[:10])
		a.ErrorIs(err, io.ErrUnexpectedEOF)
	})
	t.Run("InvalidLength", func(t *testing.T) {
		_, _, err := decodeString(bytes.Repeat([]byte{255}, 256))
		require.ErrorIs(t, err, errInvalidLength)
	})
}
