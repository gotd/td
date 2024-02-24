package crypto

import (
	"math/big"

	"github.com/go-faster/errors"
)

// CheckDHParams checks that g_a, g_b and g params meet key exchange conditions.
//
// https://core.telegram.org/mtproto/auth_key#dh-key-exchange-complete
func CheckDHParams(dhPrime, g, gA, gB *big.Int) error {
	one := big.NewInt(1)
	dhPrimeMinusOne := big.NewInt(0).Sub(dhPrime, one)
	if !InRange(g, one, dhPrimeMinusOne) {
		return errors.New("kex: bad g, g must be 1 < g < dh_prime - 1")
	}
	if !InRange(gA, one, dhPrimeMinusOne) {
		return errors.New("kex: bad g_a, g_a must be 1 < g_a < dh_prime - 1")
	}
	if !InRange(gB, one, dhPrimeMinusOne) {
		return errors.New("kex: bad g_b, g_b must be 1 < g_b < dh_prime - 1")
	}

	// IMPORTANT: Apart from the conditions on the Diffie-Hellman prime
	// dh_prime and generator g, both sides are to check that g, g_a and
	// g_b are greater than 1 and less than dh_prime - 1. We recommend
	// checking that g_a and g_b are between 2^{2048-64} and
	// dh_prime - 2^{2048-64} as well.

	// 2^{2048-64}
	safetyRangeMin := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(RSAKeyBits-64), nil)
	safetyRangeMax := big.NewInt(0).Sub(dhPrime, safetyRangeMin)
	if !InRange(gA, safetyRangeMin, safetyRangeMax) {
		return errors.New("kex: bad g_a, g_a must be 2^{2048-64} < g_a < dh_prime - 2^{2048-64}")
	}
	if !InRange(gB, safetyRangeMin, safetyRangeMax) {
		return errors.New("kex: bad g_b, g_b must be 2^{2048-64} < g_b < dh_prime - 2^{2048-64}")
	}

	return nil
}

// InRange checks whether x is in (min, max) range, i.e. min < x < max.
func InRange(x, min, max *big.Int) bool {
	return x.Cmp(min) > 0 && x.Cmp(max) < 0
}
