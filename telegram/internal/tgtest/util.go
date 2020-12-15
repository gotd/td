package tgtest

import (
	"crypto/rand"
	"crypto/sha1" // #nosec
	"encoding/hex"
	"math/big"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
)

func (s *Server) getNonceHash1(authKey, newNonce []byte) (r [16]byte) {
	var buf []byte
	buf = append(buf, newNonce...)
	buf = append(buf, 1)
	buf = append(buf, s.sha1(authKey)[0:8]...)
	copy(r[:], s.sha1(buf)[4:20])
	return
}

func (s *Server) bigFromHex(hexString string) (p *big.Int, err error) {
	data, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode hex string: %w", err)
	}

	return big.NewInt(0).SetBytes(data), nil
}

// nolint:unparam
func (s *Server) pq() (pq *big.Int, err error) {
	return big.NewInt(0x17ED48941A08F981), nil
}

func (s *Server) ga(g, dhPrime *big.Int) (a, ga *big.Int, err error) {
	one := big.NewInt(1)
	dhPrimeMinusOne := big.NewInt(0).Sub(dhPrime, one)

	safetyRangeMin := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(2048-64), nil)
	safetyRangeMax := big.NewInt(0).Sub(dhPrime, safetyRangeMin)

	randMax := big.NewInt(0).SetBit(big.NewInt(0), 2048, 1)
	for {
		a, err = rand.Int(s.rand, randMax)
		if err != nil {
			return
		}

		ga = big.NewInt(0).Exp(g, a, dhPrime)
		if crypto.InRange(ga, one, dhPrimeMinusOne) && crypto.InRange(ga, safetyRangeMin, safetyRangeMax) {
			return
		}
	}
}

func (s *Server) dhPrime() (p *big.Int, err error) {
	return s.bigFromHex("C71CAEB9C6B1C9048E6C522F70F13F73" +
		"980D40238E3E21C14934D037563D930F" +
		"48198A0AA7C14058229493D22530F4DB" +
		"FA336F6E0AC925139543AED44CCE7C37" +
		"20FD51F69458705AC68CD4FE6B6B13AB" +
		"DC9746512969328454F18FAF8C595F64" +
		"2477FE96BB2A941D5BCD1D4AC8CC4988" +
		"0708FA9B378E3C4F3A9060BEE67CF9A4" +
		"A4A695811051907E162753B56B0F6B41" +
		"0DBA74D8A84B2A14B3144E0EF1284754" +
		"FD17ED950D5965B4B9DD46582DB1178D" +
		"169C6BC465B0D6FF9CA3928FEF5B9AE4" +
		"E418FC15E83EBEA0F87FA9FF5EED7005" +
		"0DED2849F47BF959D956850CE929851F" +
		"0D8115F635B105EE2E4E15D04B2454BF" +
		"6F4FADF034B10403119CD8E3B92FCC5B")
}

func (s *Server) sha1(b []byte) []byte {
	h := sha1.New() // #nosec
	_, _ = h.Write(b)
	return h.Sum(nil)
}
