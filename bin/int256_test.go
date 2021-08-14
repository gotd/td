package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInt256_Encode(t *testing.T) {
	a := require.New(t)

	v := Int256{4, 3, 1, 2}
	b := Buffer{}
	a.NoError(v.Encode(&b))
	var decoded Int256
	a.NoError(decoded.Decode(&b))
	a.Equal(v, decoded)
	a.Error(decoded.Decode(&Buffer{}))
}

func BenchmarkBuffer_PutInt256(b *testing.B) {
	b.ReportAllocs()
	v := Int256{1, 4, 4, 6}
	buf := new(Buffer)
	for i := 0; i < b.N; i++ {
		buf.PutInt256(v)
		buf.Reset()
	}
}
