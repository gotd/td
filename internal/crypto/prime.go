package crypto

import "math/big"

// Prime checks that given number is prime.
func Prime(p *big.Int) bool {
	// TODO(tdakkota): maybe it should be smaller?
	// 1 - 1/4^64 is equal to 0.9999999999999999999999999999999999999970612641229442812300781587
	//
	// TDLib uses nchecks = 64
	// See https://github.com/tdlib/td/blob/d161323858a782bc500d188b9ae916982526c262/tdutils/td/utils/BigNum.cpp#L155.
	const probabilityN = 64

	// ProbablyPrime is mutating, so we need a copy
	return p.ProbablyPrime(probabilityN)
}
