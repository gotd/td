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
	t.Run("TooSmallBits", func(t *testing.T) {
		require.Error(t, CheckDH(3, big.NewInt(4)))
	})
}

func Test_checkPrime(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		require.NoError(t, checkPrime(big.NewInt(5)))
	})
	t.Run("PNotPrime", func(t *testing.T) {
		require.Error(t, checkPrime(big.NewInt(4)))
	})
	t.Run("HalfPMinusOneNotPrime", func(t *testing.T) {
		require.Error(t, checkPrime(big.NewInt(13)))
	})
}
