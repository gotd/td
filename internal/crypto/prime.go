package crypto

import "math/big"

func Prime(p *big.Int) bool {
	// TODO(tdakkota): maybe it should be smaller?
	// 1 - 1/4^64 is equal to 0.9999999999999999999999999999999999999970612641229442812300781587
	const probabilityN = 64

	// ProbablyPrime is mutating, so we need a copy
	cpy := big.NewInt(0).Set(p)
	return cpy.ProbablyPrime(probabilityN)
}
