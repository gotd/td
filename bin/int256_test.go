package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInt256_Encode(t *testing.T) {
	v := Int256{1043532, 12466515, 858123, 12865761}
	b := Buffer{}
	if err := v.Encode(&b); err != nil {
		t.Fatal(err)
	}
	var decoded Int256
	if err := decoded.Decode(&b); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, v, decoded)
}

func BenchmarkBuffer_PutInt256(b *testing.B) {
	b.ReportAllocs()
	v := Int256{1043532, 12466515, 858123, 12865761}
	buf := new(Buffer)
	for i := 0; i < b.N; i++ {
		buf.PutInt256(v)
		buf.Reset()
	}
}
