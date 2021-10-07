// Package srp contains implementation of Secure Remote Password protocol.
package srp

import (
	"crypto/sha512"
	"io"
	"math/big"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/xerrors"

	"github.com/gotd/xor"

	"github.com/nnqq/td/internal/crypto"
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

// Hash computes user password hash using parameters from server.
// Parameters
//
// See https://core.telegram.org/api/srp#checking-the-password-with-srp.
func (s SRP) Hash(password, srpB, random []byte, i Input) (Answer, error) {
	if err := s.checkGP(i.G, i.P); err != nil {
		return Answer{}, xerrors.Errorf("validate algo: %w", err)
	}

	p := s.bigFromBytes(i.P)
	g := big.NewInt(int64(i.G))
	gBytes, ok := s.paddedFromBig(g)
	if !ok {
		return Answer{}, xerrors.Errorf("invalid g (%d)", i.G)
	}

	// random 2048-bit number a
	a := s.bigFromBytes(random)

	// `g_a = pow(g, a) mod p`
	ga, ok := s.paddedFromBig(s.bigExp(g, a, p))
	if !ok {
		return Answer{}, xerrors.New("g_a is too big")
	}

	// `g_b = srp_B`
	gb := s.pad256(srpB)

	// `u = H(g_a | g_b)`
	u := s.bigFromBytes(s.hash(ga, gb))

	// `x = PH2(password, salt1, salt2)`
	x := s.bigFromBytes(s.secondary(password, i.Salt1, i.Salt2))

	// `v = pow(g, x) mod p`
	v := s.bigExp(g, x, p)

	// `k = (k * v) mod p`
	k := s.bigFromBytes(s.hash(i.P, gBytes))

	// `k_v = (k * v) % p`
	kv := k.Mul(k, v).Mod(k, p)

	// `t = (g_b - k_v) % p`
	t := s.bigFromBytes(srpB)
	if t.Sub(t, kv).Cmp(big.NewInt(0)) == -1 {
		t.Add(t, p)
	}

	// `s_a = pow(t, a + u * x) mod p`
	sa, ok := s.paddedFromBig(s.bigExp(t, u.Mul(u, x).Add(u, a), p))
	if !ok {
		return Answer{}, xerrors.New("s_a is too big")
	}

	// `k_a = H(s_a)`
	ka := s.hash(sa)

	// `M1 = H(H(p) xor H(g) | H2(salt1) | H2(salt2) | g_a | g_b | k_a)`
	M1 := s.hash(
		s.bytesXor(s.hash(i.P), s.hash(gBytes)),
		s.hash(i.Salt1),
		s.hash(i.Salt2),
		ga,
		gb,
		ka,
	)

	return Answer{
		A:  ga,
		M1: M1,
	}, nil
}

func (s SRP) paddedFromBig(i *big.Int) ([]byte, bool) {
	b := make([]byte, 256)
	r := crypto.FillBytes(i, b)
	return b, r
}

func (s SRP) pad256(b []byte) []byte {
	if len(b) >= 256 {
		return b[len(b)-256:]
	}

	var tmp [256]byte
	copy(tmp[256-len(b):], b)

	return tmp[:]
}

func (s SRP) bytesXor(a, b []byte) []byte {
	res := make([]byte, len(a))
	xor.Bytes(res, a, b)
	return res
}

func (s SRP) bigFromBytes(b []byte) *big.Int {
	return new(big.Int).SetBytes(b)
}

func (s SRP) bigExp(x, y, m *big.Int) *big.Int {
	return new(big.Int).Exp(x, y, m)
}

// The main hashing function H is sha256:
//
// H(data) := sha256(data)
func (s SRP) hash(data ...[]byte) []byte {
	return crypto.SHA256(data...)
}

// The salting hashing function SH is defined as follows:
//
// SH(data, salt) := H(salt | data | salt)
func (s SRP) saltHash(data, salt []byte) []byte {
	return s.hash(salt, data, salt)
}

// The primary password hashing function is defined as follows:
//
// PH1(password, salt1, salt2) := SH(SH(password, salt1), salt2)
func (s SRP) primary(password, salt1, salt2 []byte) []byte {
	return s.saltHash(s.saltHash(password, salt1), salt2)
}

// The secondary password hashing function is defined as follows:
//
// PH2(password, salt1, salt2) := SH(pbkdf2(sha512, PH1(password, salt1, salt2), salt1, 100000), salt2)
func (s SRP) secondary(password, salt1, salt2 []byte) []byte {
	return s.saltHash(s.pbkdf2(s.primary(password, salt1, salt2), salt1, 100000), salt2)
}

func (s SRP) pbkdf2(ph1, salt1 []byte, n int) []byte {
	return pbkdf2.Key(ph1, salt1, n, 64, sha512.New)
}

func (s SRP) checkGP(g int, pBytes []byte) error {
	p := s.bigFromBytes(pBytes)
	return crypto.CheckDH(g, p)
}
