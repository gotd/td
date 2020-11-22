package bin

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInt128_Encode(t *testing.T) {
	v := Int128{1043532, 12466515}
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

func TestInt128_BigInt(t *testing.T) {
	v := Int128{}
	v.SetToBigInt(big.NewInt(0x17ED48941A08F981))
	b := Buffer{}
	if err := v.Encode(&b); err != nil {
		t.Fatal(err)
	}
	var decoded Int128
	if err := decoded.Decode(&b); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, v, decoded)
	require.Zero(t, decoded.BigInt().Cmp(big.NewInt(0x17ED48941A08F981)))

	t.Run("32Bit", func(t *testing.T) {
		vb := Int128{}
		if err := vb.setToBigInt32([]big.Word{1, 2, 10, 15}); err != nil {
			t.Fatal(err)
		}
	})
}

func BenchmarkBuffer_PutInt128(b *testing.B) {
	v := Int128{10, 15}
	buf := new(Buffer)
	for i := 0; i < b.N; i++ {
		buf.PutInt128(v)
		buf.Reset()
	}
}
