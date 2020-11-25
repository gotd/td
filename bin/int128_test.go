package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInt128_Encode(t *testing.T) {
	v := Int128{1, 2, 3, 0, 134, 45}
	b := Buffer{}
	if err := v.Encode(&b); err != nil {
		t.Fatal(err)
	}
	var decoded Int128
	if err := decoded.Decode(&b); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, v, decoded)
}

func BenchmarkBuffer_PutInt128(b *testing.B) {
	v := Int128{10, 15}
	buf := new(Buffer)
	for i := 0; i < b.N; i++ {
		buf.PutInt128(v)
		buf.Reset()
	}
}
