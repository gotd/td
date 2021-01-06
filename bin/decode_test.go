package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecoder_ID(t *testing.T) {
	var b Buffer
	const id = 0x1234
	b.PutID(id)
	b.PutString("foo bar")
	b.PutBool(true)
	b.PutBytes([]byte{1, 2, 3, 4})
	b.PutInt32(-150)
	b.PutInt(-151)
	b.PutLong(100)
	b.PutDouble(1.5)
	b.PutVectorHeader(10)

	a := require.New(t)

	gotID, err := b.ID()
	a.NoError(err)
	a.Equal(uint32(id), gotID)
	gotStr, err := b.String()
	a.NoError(err)
	a.Equal("foo bar", gotStr)
	gotBool, err := b.Bool()
	a.NoError(err)
	a.True(gotBool)
	gotBytes, err := b.Bytes()
	a.NoError(err)
	a.Equal([]byte{1, 2, 3, 4}, gotBytes)
	gotInt32, err := b.Int32()
	a.NoError(err)
	a.Equal(int32(-150), gotInt32)
	gotInt, err := b.Int()
	a.NoError(err)
	a.Equal(-151, gotInt)
	gotLong, err := b.Long()
	a.NoError(err)
	a.Equal(int64(100), gotLong)
	gotDouble, err := b.Double()
	a.NoError(err)
	a.Equal(1.5, gotDouble)
	gotVectorHead, err := b.VectorHeader()
	a.NoError(err)
	a.Equal(10, gotVectorHead)
	require.Zero(t, b.Len(), "buffer should be fully consumed")
}

func TestConsumePeek(t *testing.T) {
	a := require.New(t)
	b := Buffer{}
	b.PutID(0x1)

	var buf [4]byte
	err := b.PeekN(buf[:], len(buf))
	a.NoError(err)
	a.Equal([...]byte{1, 0, 0, 0}, buf)
	// Check that peek does not increase cursor.
	a.Equal([]byte{1, 0, 0, 0}, b.Buf)

	id, err := b.PeekID()
	a.NoError(err)
	a.Equal(uint32(0x1), id)
	// Check that peek does not increase cursor.
	a.Equal([]byte{1, 0, 0, 0}, b.Buf)

	err = b.ConsumeN(buf[:], len(buf))
	a.NoError(err)
	// Check that consume increase cursor.
	a.Len(b.Buf, 0)
}
