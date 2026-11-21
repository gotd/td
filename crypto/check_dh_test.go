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

// withoutSafePrimes empties safePrimes for the test and restores it after.
func withoutSafePrimes(t *testing.T) {
	safePrimes.mux.Lock()
	saved := safePrimes.primes
	safePrimes.primes = nil
	safePrimes.mux.Unlock()

	t.Cleanup(func() {
		safePrimes.mux.Lock()
		safePrimes.primes = saved
		safePrimes.mux.Unlock()
	})
}

func rememberedSafePrimes() int {
	safePrimes.mux.RLock()
	defer safePrimes.mux.RUnlock()
	return len(safePrimes.primes)
}

func TestCheckDHRemembersSafePrime(t *testing.T) {
	withoutSafePrimes(t)
	a := require.New(t)

	p := new(big.Int).Set(checkGPdhPrime)
	a.False(isSafePrime(p))
	a.NoError(CheckDH(checkGPg, p))
	a.True(isSafePrime(checkGPdhPrime))
	a.NoError(CheckDH(checkGPg, p))
	a.Equal(1, rememberedSafePrimes())

	// The remembered prime is a copy of p.
	p.Add(p, big.NewInt(2))
	a.True(isSafePrime(checkGPdhPrime))
	a.False(isSafePrime(p))

	// -p has the same bytes but is not prime.
	a.Error(checkPrime(new(big.Int).Neg(checkGPdhPrime)))
}

func Test_checkPrimeRemembersOnlySafePrimes(t *testing.T) {
	withoutSafePrimes(t)
	a := require.New(t)

	a.Error(checkPrime(big.NewInt(4)))
	a.Error(checkPrime(big.NewInt(13)))
	a.Zero(rememberedSafePrimes())

	// Safe primes: p and (p-1)/2 are both prime. There are more of them
	// than maxSafePrimes, and each passes every time.
	for _, p := range []int64{5, 7, 11, 23, 47, 59, 83, 107, 167, 179} {
		a.NoError(checkPrime(big.NewInt(p)))
		a.NoError(checkPrime(big.NewInt(p)))
	}
	a.Equal(maxSafePrimes, rememberedSafePrimes())
}
