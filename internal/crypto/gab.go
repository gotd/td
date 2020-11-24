package crypto

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"
)

// GAB is part of Diffie-Hellman key exchange.
//
// See https://en.wikipedia.org/wiki/Diffie%E2%80%93Hellman_key_exchange
//
// Values:
//	gA is g_a
//	dhPrime is dh_prime
//	gB is g_b
//	gAB is gab.
func GAB(dhPrime, g, gA *big.Int, randSource io.Reader) (b, gB, gAB *big.Int, err error) {
	randMax := big.NewInt(0).SetBit(big.NewInt(0), 2048, 1)
	if b, err = rand.Int(randSource, randMax); err != nil {
		return nil, nil, nil, err
	}
	gB, gAB = gab(dhPrime, b, g, gA)
	return b, gB, gAB, nil
}

func gab(dhPrime, b, g, gA *big.Int) (gB, gAB *big.Int) {
	gB = big.NewInt(0).Exp(g, b, dhPrime)
	gAB = big.NewInt(0).Exp(gA, b, dhPrime)
	return gB, gAB
}

// CheckGAB checks that g_a, g_b and g params meet key exchange conditions.
//
// https://core.telegram.org/mtproto/auth_key#dh-key-exchange-complete
func CheckGAB(dhPrime, g, gA, gB *big.Int) error {
	one := big.NewInt(1)
	dhPrimeMinusOne := big.NewInt(0).Sub(dhPrime, one)
	if !inRange(g, one, dhPrimeMinusOne) {
		return errors.New("kex: bad g")
	}
	if !inRange(gA, one, dhPrimeMinusOne) {
		return errors.New("kex: bad g_a")
	}
	if !inRange(gB, one, dhPrimeMinusOne) {
		return errors.New("kex: bad g_b")
	}

	// IMPORTANT: Apart from the conditions on the Diffie-Hellman prime
	// dh_prime and generator g, both sides are to check that g, g_a and
	// g_b are greater than 1 and less than dh_prime - 1. We recommend
	// checking that g_a and g_b are between 2^{2048-64} and
	// dh_prime - 2^{2048-64} as well.

	// 2^{2048-64}
	safetyRangeMin := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(2048-64), nil)
	safetyRangeMax := big.NewInt(0).Sub(dhPrime, safetyRangeMin)
	if !inRange(gA, safetyRangeMin, safetyRangeMax) {
		return errors.New("kex: bad g_a")
	}
	if !inRange(gB, safetyRangeMin, safetyRangeMax) {
		return errors.New("kex: bad g_b")
	}

	return nil
}

// inRange checks whether x is in (min, max) range, i.e. min < x < max.
func inRange(x, min, max *big.Int) bool {
	return x.Cmp(min) > 0 && x.Cmp(max) < 0
}
