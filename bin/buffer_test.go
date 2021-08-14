package bin

import (
	"errors"
	"io"
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

func TestBuffer_Raw(t *testing.T) {
	b := Buffer{Buf: []byte{1, 2, 3}}
	b.Raw()[0] = 2
	require.Equal(t, []byte{2, 2, 3}, b.Buf)
}

func TestBuffer_ResetTo(t *testing.T) {
	orig := []byte{1, 2, 3}
	b := Buffer{Buf: orig}

	b.ResetTo([]byte{4, 5, 6})
	b.Buf[0] = 2

	require.Equal(t, []byte{1, 2, 3}, orig)
	require.Equal(t, []byte{2, 5, 6}, b.Buf)
}

var testErr = errors.New("test")

type errObject struct{}

func (e errObject) Encode(b *Buffer) error {
	return testErr
}

func (e errObject) Decode(b *Buffer) error {
	return testErr
}

type bytesObject uint32

func (o bytesObject) Decode(b *Buffer) error {
	return b.ConsumeID(uint32(o))
}

func (o bytesObject) Encode(b *Buffer) error {
	b.PutUint32(uint32(o))
	return nil
}

func TestBufferEncodeDecode(t *testing.T) {
	a := require.New(t)
	b := Buffer{}
	a.ErrorIs(b.Encode(errObject{}), testErr)
	a.ErrorIs(b.Decode(errObject{}), testErr)
	a.NoError(b.Encode(bytesObject(1)))
	a.NoError(b.Decode(bytesObject(1)))
}

func TestBuffer_Read(t *testing.T) {
	a := require.New(t)
	b := Buffer{}

	_, err := b.Read(nil)
	a.NoError(err)

	_, err = b.Read([]byte{1})
	a.ErrorIs(err, io.EOF)
	a.Equal(err, io.EOF)

	b.Put([]byte{1})
	_, err = b.Read([]byte{1})
	a.NoError(err)
	a.Empty(b.Buf)
}
