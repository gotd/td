// Package srp contains implementation of Secure Remote Password protocol.
package srp

import (
	"crypto/sha256"
	"io"
	"math/big"

	"github.com/go-faster/xor"

	"github.com/gotd/td/crypto"
)

// SRP is client implementation of Secure Remote Password protocol.
//
// See https://core.telegram.org/api/srp.
type SRP struct {
	random io.Reader
}

// NewSRP creates new SRP instance.
func NewSRP(random io.Reader) SRP {
	return SRP{random: random}
}

// Input is hashing algorithm parameters from server.
//
// Copy of tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow
type Input struct {
	// One of two salts used by the derivation function (see SRP 2FA login)
	Salt1 []byte
	// One of two salts used by the derivation function (see SRP 2FA login)
	Salt2 []byte
	// Base (see SRP 2FA login)
	G int
	// 2048-bit modulus (see SRP 2FA login)
	P []byte
}

// Answer is result of SRP algorithm.
type Answer struct {
	// A parameter (see SRP)
	A []byte
	// M1 parameter (see SRP)
	M1 []byte
}

func xor32(a, b [sha256.Size]byte) (res [sha256.Size]byte) {
	xor.Bytes(res[:], a[:], b[:])
	return res
}

func (s SRP) bigFromBytes(b []byte) *big.Int {
	return new(big.Int).SetBytes(b)
}

func (s SRP) bigExp(x, y, m *big.Int) *big.Int {
	return new(big.Int).Exp(x, y, m)
}

func checkInput(g int, p *big.Int) error {
	return crypto.CheckDH(g, p)
}
