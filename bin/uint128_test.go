package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUint128_Encode(t *testing.T) {
	v := Uint128{1043532, 12466515}
	b := Buffer{}
	if err := v.Encode(&b); err != nil {
		t.Fatal(err)
	}
	var decoded Uint128
	if err := decoded.Decode(&b); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, v, decoded)
}

func BenchmarkBuffer_PutUint128(b *testing.B) {
	v := Uint128{10, 15}
	buf := new(Buffer)
	for i := 0; i < b.N; i++ {
		buf.PutUint128(v)
		buf.Reset()
	}
}
