package bin

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/testutil"
)

func TestExpandReset(t *testing.T) {
	a := require.New(t)
	b := Buffer{}
	b.PutInt(10)
	before := b.Len()
	copyBuf := b.Copy()

	b.Expand(2)
	a.Equal(before+2, b.Len())
	a.Equal(copyBuf, b.Buf[:before], "buffer overwrite")

	b.ResetN(b.Len() + 2)
	a.Zero(b.Buf[0], "buffer not zeroed")

	before = b.Len()
	b.Skip(2)
	a.Equal(before-2, b.Len())
}

func TestCopy(t *testing.T) {
	b := Buffer{}
	b.PutInt(10)
	copyBuf := b.Copy()
	copyBuf[0] = 1
	require.Equal(t, byte(10), b.Buf[0], "buffer overwritten from copy")
}

func TestBuffer_ResetN(t *testing.T) {
	var b Buffer
	testutil.ZeroAlloc(t, func() {
		b.ResetN(1024)
	})
}
