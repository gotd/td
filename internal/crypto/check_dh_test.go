package crypto

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkCheckDH(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = CheckDH(checkGPg, checkGPdhPrime)
	}
}

func TestCheckDH(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		a := require.New(t)
		a.NoError(CheckDH(checkGPg, checkGPdhPrime))
	})
	t.Run("WrongG", func(t *testing.T) {
		require.Error(t, CheckDH(1337, checkGPdhPrime))
	})
	t.Run("WrongDivider", func(t *testing.T) {
		// CheckGP should check that p mod 3 = 2 for g = 3;
		// We pass p = 4, so p mod 3 = 1.
		require.Error(t, CheckDH(3, big.NewInt(4)))
	})
}
