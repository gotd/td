package bin

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuffer_PeekN(t *testing.T) {
	a := require.New(t)
	b := Buffer{Buf: []byte{1, 2, 3}}

	target := make([]byte, 4)
	a.ErrorIs(b.PeekN(target, 4), io.ErrUnexpectedEOF)
	a.NoError(b.PeekN(target, 3))
	a.Equal([]byte{1, 2, 3, 0}, target)
	a.Equal([]byte{1, 2, 3}, b.Buf, "PeekN must not modify buffer")
}

func TestBuffer_ConsumeN(t *testing.T) {
	a := require.New(t)
	b := Buffer{Buf: []byte{1, 2, 3}}

	target := make([]byte, 4)
	a.ErrorIs(b.ConsumeN(target, 4), io.ErrUnexpectedEOF)
	a.NoError(b.ConsumeN(target, 3))
	a.Equal([]byte{1, 2, 3, 0}, target)
	a.Empty(b.Buf, "ConsumeN must modify buffer")
}

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

func TestBuffer_VectorHeader(t *testing.T) {
	tests := []struct {
		data        []byte
		expected    int
		wantErr     bool
		targetError error
	}{
		{nil, 0, true, io.ErrUnexpectedEOF},
		{typeIDToBytes(TypeVector), 0, true, io.ErrUnexpectedEOF},
		{typeIDToBytes(TypeFalse), 0, true, nil},
		{append(typeIDToBytes(TypeVector), 0, 0, 0, 0), 0, false, nil},
		{append(typeIDToBytes(TypeVector), 255, 255, 255, 255), 0, true, nil},
	}
	for _, tt := range tests {
		a := require.New(t)
		b := &Buffer{Buf: tt.data}

		r, err := b.VectorHeader()
		if tt.wantErr {
			a.Error(err)
			if tt.targetError != nil {
				a.ErrorIs(err, tt.targetError)
			}
			a.Zero(r)
		} else {
			a.NoError(err)
			a.Equal(tt.expected, r)
		}
	}
}

func TestBuffer_Int(t *testing.T) {
	tests := []struct {
		data        []byte
		expected    int
		wantErr     bool
		targetError error
	}{
		{nil, 0, true, io.ErrUnexpectedEOF},
		{make([]byte, 3), 0, true, io.ErrUnexpectedEOF},
		{make([]byte, 4), 0, false, nil},
		{[]byte{0x01, 0x00, 0x00, 0x00}, 1, false, nil},
	}
	for _, tt := range tests {
		a := require.New(t)
		b := &Buffer{Buf: tt.data}

		r, err := b.Int()
		if tt.wantErr {
			a.Error(err)
			if tt.targetError != nil {
				a.ErrorIs(err, tt.targetError)
			}
			a.Zero(r)
		} else {
			a.NoError(err)
			a.Equal(tt.expected, r)
		}
	}
}

func TestBuffer_Double(t *testing.T) {
	tests := []struct {
		data        []byte
		expected    float64
		wantErr     bool
		targetError error
	}{
		{nil, 0, true, io.ErrUnexpectedEOF},
		{make([]byte, 7), 0, true, io.ErrUnexpectedEOF},
		{make([]byte, 8), 0, false, nil},
		{func() []byte {
			d := Buffer{}
			d.PutDouble(1)
			return d.Buf
		}(), 1, false, nil},
	}
	for _, tt := range tests {
		a := require.New(t)
		b := &Buffer{Buf: tt.data}

		r, err := b.Double()
		if tt.wantErr {
			a.Error(err)
			if tt.targetError != nil {
				a.ErrorIs(err, tt.targetError)
			}
			a.Zero(r)
		} else {
			a.NoError(err)
			a.Equal(tt.expected, r)
		}
	}
}

func TestBuffer_Bool(t *testing.T) {
	tests := []struct {
		data     []byte
		expected bool
		wantErr  bool
	}{
		{typeIDToBytes(TypeTrue), true, false},
		{typeIDToBytes(TypeFalse), false, false},
		{nil, false, true},
		{typeIDToBytes(TypeVector), false, true},
	}
	for _, tt := range tests {
		a := require.New(t)
		b := &Buffer{Buf: tt.data}

		r, err := b.Bool()
		if tt.wantErr {
			a.Error(err)
			a.Zero(r)
		} else {
			a.NoError(err)
			a.Equal(tt.expected, r)
		}
	}
}

func TestBuffer_ConsumeID(t *testing.T) {
	consume := uint32(TypeTrue)
	tests := []struct {
		data        []byte
		wantErr     bool
		targetError func(a *require.Assertions, e error)
	}{
		{typeIDToBytes(TypeTrue), false, nil},
		{
			typeIDToBytes(TypeVector),
			true,
			func(a *require.Assertions, e error) {
				var unexpected *UnexpectedIDErr
				a.ErrorAs(e, &unexpected)
				a.Equal(uint32(TypeVector), unexpected.ID)
				a.NotEmpty(unexpected.Error())
			},
		},
		{nil, true, nil},
	}
	for _, tt := range tests {
		a := require.New(t)
		b := &Buffer{Buf: tt.data}

		err := b.ConsumeID(consume)
		if tt.wantErr {
			a.Error(err)
			if tt.targetError != nil {
				tt.targetError(a, err)
			}
		} else {
			a.NoError(err)
		}
	}
}

func TestBuffer_String(t *testing.T) {
	b := Buffer{}
	_, err := b.String()
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	b.PutString("ab")
	b.Buf = b.Buf[:3] // Cut padding

	_, err = b.String()
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
}

func TestBuffer_Bytes(t *testing.T) {
	b := Buffer{}
	_, err := b.Bytes()
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	b.PutBytes([]byte("ab"))
	b.Buf = b.Buf[:3] // Cut padding

	_, err = b.Bytes()
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
}
