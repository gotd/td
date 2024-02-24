package exchange

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"math/big"

	"github.com/go-faster/errors"

	"github.com/gotd/td/crypto"
)

// ServerRNG is server-side random number generator.
type ServerRNG interface {
	PQ() (pq *big.Int, err error)
	GA(g int, dhPrime *big.Int) (a, ga *big.Int, err error)
	DhPrime() (p *big.Int, err error)
}

var _ ServerRNG = TestServerRNG{}

// TestServerRNG implements testing-only ServerRNG.
type TestServerRNG struct {
	rand io.Reader
}

func (s TestServerRNG) bigFromHex(hexString string) (p *big.Int, err error) {
	data, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, errors.Wrap(err, "decode hex string")
	}

	return big.NewInt(0).SetBytes(data), nil
}

// PQ always returns testing pq value.
//
// nolint:unparam
func (s TestServerRNG) PQ() (pq *big.Int, err error) {
	return big.NewInt(0x17ED48941A08F981), nil
}

// GA returns testing a and g_a params.
func (s TestServerRNG) GA(g int, dhPrime *big.Int) (a, ga *big.Int, err error) {
	if err := crypto.CheckGP(g, dhPrime); err != nil {
		return nil, nil, err
	}

	gBig := big.NewInt(int64(g))
	one := big.NewInt(1)
	dhPrimeMinusOne := big.NewInt(0).Sub(dhPrime, one)

	safetyRangeMin := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(crypto.RSAKeyBits-64), nil)
	safetyRangeMax := big.NewInt(0).Sub(dhPrime, safetyRangeMin)

	randMax := big.NewInt(0).SetBit(big.NewInt(0), crypto.RSAKeyBits, 1)
	for {
		a, err = rand.Int(s.rand, randMax)
		if err != nil {
			return
		}

		ga = big.NewInt(0).Exp(gBig, a, dhPrime)
		if crypto.InRange(ga, one, dhPrimeMinusOne) && crypto.InRange(ga, safetyRangeMin, safetyRangeMax) {
			return
		}
	}
}

// DhPrime always returns testing dh_prime.
func (s TestServerRNG) DhPrime() (p *big.Int, err error) {
	return s.bigFromHex("C71CAEB9C6B1C9048E6C522F70F13F73980D40238E3E21C14934D037563D930F" +
		"48198A0AA7C14058229493D22530F4DBFA336F6E0AC925139543AED44CCE7C37" +
		"20FD51F69458705AC68CD4FE6B6B13ABDC9746512969328454F18FAF8C595F64" +
		"2477FE96BB2A941D5BCD1D4AC8CC49880708FA9B378E3C4F3A9060BEE67CF9A4" +
		"A4A695811051907E162753B56B0F6B410DBA74D8A84B2A14B3144E0EF1284754" +
		"FD17ED950D5965B4B9DD46582DB1178D169C6BC465B0D6FF9CA3928FEF5B9AE4" +
		"E418FC15E83EBEA0F87FA9FF5EED70050DED2849F47BF959D956850CE929851F" +
		"0D8115F635B105EE2E4E15D04B2454BF6F4FADF034B10403119CD8E3B92FCC5B")
}
