package exchange

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"math/big"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/proto"

	"github.com/gotd/td/internal/crypto"
)

type ServerExchange struct {
	unencryptedWriter
	key *rsa.PrivateKey
}

func NewServerExchange(c Config, key *rsa.PrivateKey) *ServerExchange {
	return &ServerExchange{
		unencryptedWriter: unencryptedWriter{
			Config: c,
			input:  proto.MessageFromClient,
			output: proto.MessageServerResponse,
		},
		key: key,
	}
}

type ServerExchangeResult struct {
	Key        crypto.AuthKeyWithID
	ServerSalt int64
}

func (s *ServerExchange) bigFromHex(hexString string) (p *big.Int, err error) {
	data, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode hex string: %w", err)
	}

	return big.NewInt(0).SetBytes(data), nil
}

// nolint:unparam
func (s *ServerExchange) pq() (pq *big.Int, err error) {
	return big.NewInt(0x17ED48941A08F981), nil
}

func (s *ServerExchange) ga(g int, dhPrime *big.Int) (a, ga *big.Int, err error) {
	if err := crypto.CheckGP(g, dhPrime); err != nil {
		return nil, nil, err
	}

	gBig := big.NewInt(int64(g))
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

		ga = big.NewInt(0).Exp(gBig, a, dhPrime)
		if crypto.InRange(ga, one, dhPrimeMinusOne) && crypto.InRange(ga, safetyRangeMin, safetyRangeMax) {
			return
		}
	}
}

func (s *ServerExchange) dhPrime() (p *big.Int, err error) {
	return s.bigFromHex("C71CAEB9C6B1C9048E6C522F70F13F73980D40238E3E21C14934D037563D930F" +
		"48198A0AA7C14058229493D22530F4DBFA336F6E0AC925139543AED44CCE7C37" +
		"20FD51F69458705AC68CD4FE6B6B13ABDC9746512969328454F18FAF8C595F64" +
		"2477FE96BB2A941D5BCD1D4AC8CC49880708FA9B378E3C4F3A9060BEE67CF9A4" +
		"A4A695811051907E162753B56B0F6B410DBA74D8A84B2A14B3144E0EF1284754" +
		"FD17ED950D5965B4B9DD46582DB1178D169C6BC465B0D6FF9CA3928FEF5B9AE4" +
		"E418FC15E83EBEA0F87FA9FF5EED70050DED2849F47BF959D956850CE929851F" +
		"0D8115F635B105EE2E4E15D04B2454BF6F4FADF034B10403119CD8E3B92FCC5B")
}
