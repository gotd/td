package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUint256_Encode(t *testing.T) {
	v := Uint256{1043532, 12466515, 858123, 12865761}
	b := Buffer{}
	if err := v.Encode(&b); err != nil {
		t.Fatal(err)
	}
	var decoded Uint256
	if err := decoded.Decode(&b); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, v, decoded)
}

func BenchmarkBuffer_PutUint256(b *testing.B) {
	b.ReportAllocs()
	v := Uint256{1043532, 12466515, 858123, 12865761}
	buf := new(Buffer)
	for i := 0; i < b.N; i++ {
		buf.PutUint256(v)
		buf.Reset()
	}
}
