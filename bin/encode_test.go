package bin

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkBuffer_PutString(b *testing.B) {
	b.ReportAllocs()
	buf := new(Buffer)
	for i := 0; i < b.N; i++ {
		buf.PutString("Foo bar baz")
		buf.Reset()
	}
}

func BenchmarkBuffer_PutID(b *testing.B) {
	b.ReportAllocs()
	buf := new(Buffer)
	for i := 0; i < b.N; i++ {
		buf.PutID(TypeStringID)
		buf.Reset()
	}
}

func BenchmarkBufferMultiplePuts(b *testing.B) {
	b.ReportAllocs()
	buf := new(Buffer)
	for i := 0; i < b.N; i++ {
		buf.PutID(0xbadbad)
		buf.PutBool(true)
		buf.PutString("foo")
		buf.PutLong(12345)
		buf.PutDouble(10.55)
		buf.PutInt(10)
		buf.PutVectorHeader(2)
		buf.PutInt(1)
		buf.PutInt(2)
		buf.Reset()
	}
}

func BenchmarkBuffer_Put(b *testing.B) {
	b.ReportAllocs()
	buf := new(Buffer)
	for i := 0; i < b.N; i++ {
		raw := []byte{1, 2, 3}
		buf.PutID(0xbadbad)
		buf.Put(raw)
		buf.Reset()
	}
}

func TestBuffer_PutInt32(t *testing.T) {
	for _, tt := range []struct {
		Integer int32
		Value   []byte
	}{
		{Integer: 0, Value: []byte{0x00, 0x00, 0x00, 0x00}},
		{Integer: 1, Value: []byte{0x01, 0x00, 0x00, 0x00}},
		{Integer: -1, Value: []byte{0xff, 0xff, 0xff, 0xff}},
		{Integer: math.MaxInt32, Value: []byte{0xff, 0xff, 0xff, 0x7f}},
		{Integer: math.MinInt32, Value: []byte{0x00, 0x00, 0x00, 0x80}},
	} {
		t.Run(fmt.Sprintf("%d", tt.Integer), func(t *testing.T) {
			var b Buffer
			b.PutInt32(tt.Integer)
			require.Equal(t, tt.Value, b.Buf)

			t.Run("Int", func(t *testing.T) {
				b.Reset()
				b.PutInt(int(tt.Integer))
				require.Equal(t, tt.Value, b.Buf)
			})
		})
	}
}

func TestBuffer_PutUint32(t *testing.T) {
	for _, tt := range []struct {
		Integer uint32
		Value   []byte
	}{
		{Integer: 0, Value: []byte{0x00, 0x00, 0x00, 0x00}},
		{Integer: 1, Value: []byte{0x01, 0x00, 0x00, 0x00}},
		{Integer: math.MaxUint32, Value: []byte{0xff, 0xff, 0xff, 0xff}},
	} {
		t.Run(fmt.Sprintf("%d", tt.Integer), func(t *testing.T) {
			var b Buffer
			b.PutUint32(tt.Integer)
			require.Equal(t, tt.Value, b.Buf)
		})
	}
}

func TestBuffer_PutLong(t *testing.T) {
	for _, tt := range []struct {
		Integer int64
		Value   []byte
	}{
		{Integer: 0, Value: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		{Integer: 1, Value: []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		{Integer: -1, Value: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{Integer: math.MaxInt64, Value: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{Integer: math.MinInt64, Value: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x80}},
	} {
		t.Run(fmt.Sprintf("%d", tt.Integer), func(t *testing.T) {
			var b Buffer
			b.PutLong(tt.Integer)
			require.Equal(t, tt.Value, b.Buf)
		})
	}
}

func TestBuffer_PutDouble(t *testing.T) {
	for _, tt := range []struct {
		Float float64
		Value []byte
	}{
		{Float: 0, Value: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		{Float: 1.5, Value: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf8, 0x3f}},
		{Float: -1.5, Value: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf8, 0xbf}},
		{Float: math.Inf(1), Value: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0x7f}},
		{Float: math.Inf(-1), Value: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf0, 0xff}},
	} {
		t.Run(fmt.Sprintf("%f", tt.Float), func(t *testing.T) {
			var b Buffer
			b.PutDouble(tt.Float)
			require.Equal(t, tt.Value, b.Buf)
		})
	}
}

func typeIDToBytes(id uint32) []byte {
	b := Buffer{}
	b.PutID(id)
	return b.Buf
}

func TestBuffer_PutBool(t *testing.T) {
	for _, tt := range []struct {
		Bool  bool
		Value []byte
	}{
		{Bool: true, Value: typeIDToBytes(TypeTrue)},
		{Bool: false, Value: typeIDToBytes(TypeFalse)},
	} {
		t.Run(fmt.Sprintf("%t", tt.Bool), func(t *testing.T) {
			var b Buffer
			b.PutBool(tt.Bool)
			require.Equal(t, tt.Value, b.Buf)
		})
	}
}

func TestBuffer_PutUint16(t *testing.T) {
	for _, tt := range []struct {
		Integer uint16
		Value   []byte
	}{
		{Integer: 0, Value: []byte{0x00, 0x00}},
		{Integer: 1, Value: []byte{0x01, 0x00}},
		{Integer: math.MaxUint16, Value: []byte{0xff, 0xff}},
	} {
		t.Run(fmt.Sprintf("%d", tt.Integer), func(t *testing.T) {
			var b Buffer
			b.PutUint16(tt.Integer)
			require.Equal(t, tt.Value, b.Buf)
		})
	}
}

func TestBuffer_PutVectorHeader(t *testing.T) {
	for _, tt := range []struct {
		Len   int
		Value []byte
	}{
		{
			Len:   0,
			Value: []byte{0x15, 0xc4, 0xb5, 0x1c, 0x0, 0x0, 0x0, 0x0},
		},
	} {
		t.Run(fmt.Sprintf("Vec[%d]", tt.Len), func(t *testing.T) {
			var b Buffer
			b.PutVectorHeader(tt.Len)
			require.Equal(t, tt.Value, b.Buf)
		})
	}
}

func TestBuffer_Put(t *testing.T) {
	a := require.New(t)
	b := Buffer{Buf: []byte{1, 2, 3}}

	b.Put(nil)
	a.Len(b.Buf, 3)

	b.Put([]byte{})
	a.Len(b.Buf, 3)

	b.Put([]byte{4})
	a.Equal([]byte{1, 2, 3, 4}, b.Buf)
}
