package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInt128_Encode(t *testing.T) {
	a := require.New(t)

	v := Int128{1, 2, 3, 0, 134, 45}
	b := Buffer{}
	a.NoError(v.Encode(&b))
	var decoded Int128
	a.NoError(decoded.Decode(&b))
	a.Equal(v, decoded)
	a.Error(decoded.Decode(&Buffer{}))
}

func BenchmarkBuffer_PutInt128(b *testing.B) {
	v := Int128{10, 15}
	buf := new(Buffer)
	for i := 0; i < b.N; i++ {
		buf.PutInt128(v)
		buf.Reset()
	}
}
