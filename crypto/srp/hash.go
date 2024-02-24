package srp

import (
	"crypto/sha256"
	"crypto/sha512"
	"math/big"

	"github.com/go-faster/errors"
	"golang.org/x/crypto/pbkdf2"
)

// Hash computes user password hash using parameters from server.
//
// See https://core.telegram.org/api/srp#checking-the-password-with-srp.
func (s SRP) Hash(password, srpB, random []byte, i Input) (Answer, error) {
	p := s.bigFromBytes(i.P)
	if err := checkInput(i.G, p); err != nil {
		return Answer{}, errors.Wrap(err, "validate algo")
	}

	g := big.NewInt(int64(i.G))
	// It is safe to use FillBytes directly because we know that 64-bit G always smaller than
	// 256-bit destination array.
	var gBytes [256]byte
	g.FillBytes(gBytes[:])

	// random 2048-bit number a
	a := s.bigFromBytes(random)

	// `g_a = pow(g, a) mod p`
	ga, ok := s.pad256FromBig(s.bigExp(g, a, p))
	if !ok {
		return Answer{}, errors.New("g_a is too big")
	}

	// `g_b = srp_B`
	gb := s.pad256(srpB)

	// `u = H(g_a | g_b)`
	u := s.bigFromBytes(s.hash(ga[:], gb[:]))

	// `x = PH2(password, salt1, salt2)`
	// `v = pow(g, x) mod p`
	x, v := s.computeXV(password, i.Salt1, i.Salt2, g, p)

	// `k = (k * v) mod p`
	k := s.bigFromBytes(s.hash(i.P, gBytes[:]))

	// `k_v = (k * v) % p`
	kv := k.Mul(k, v).Mod(k, p)

	// `t = (g_b - k_v) % p`
	t := s.bigFromBytes(srpB)
	if t.Sub(t, kv).Cmp(big.NewInt(0)) == -1 {
		t.Add(t, p)
	}

	// `s_a = pow(t, a + u * x) mod p`
	sa, ok := s.pad256FromBig(s.bigExp(t, u.Mul(u, x).Add(u, a), p))
	if !ok {
		return Answer{}, errors.New("s_a is too big")
	}

	// `k_a = H(s_a)`
	ka := sha256.Sum256(sa[:])

	// `M1 = H(H(p) xor H(g) | H2(salt1) | H2(salt2) | g_a | g_b | k_a)`
	xorHpHg := xor32(sha256.Sum256(i.P), sha256.Sum256(gBytes[:]))
	M1 := s.hash(
		xorHpHg[:],
		s.hash(i.Salt1),
		s.hash(i.Salt2),
		ga[:],
		gb[:],
		ka[:],
	)

	return Answer{
		A:  ga[:],
		M1: M1,
	}, nil
}

// The main hashing function H is sha256:
//
// H(data) := sha256(data)
func (s SRP) hash(data ...[]byte) []byte {
	h := sha256.New()
	for i := range data {
		h.Write(data[i])
	}
	return h.Sum(nil)
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
