package crypto

import (
	"math/big"

	"github.com/go-faster/errors"
)

// CheckDH performs DH parameters check described in Telegram docs.
//
//	Client is expected to check whether p is a safe 2048-bit prime (meaning that both p and (p-1)/2 are prime,
//	and that 2^2047 < p < 2^2048), and that g generates a cyclic subgroup of prime order (p-1)/2, i.e.
//	is a quadratic residue mod p. Since g is always equal to 2, 3, 4, 5, 6 or 7, this is easily done using quadratic
//	reciprocity law, yielding a simple condition on p mod 4g — namely, p mod 8 = 7 for g = 2; p mod 3 = 2 for g = 3;
//	no extra condition for g = 4; p mod 5 = 1 or 4 for g = 5; p mod 24 = 19 or 23 for g = 6; and p mod 7 = 3,
//	5 or 6 for g = 7.
//
// See https://core.telegram.org/mtproto/auth_key#presenting-proof-of-work-server-authentication.
//
// See https://core.telegram.org/api/srp#checking-the-password-with-srp.
//
// See https://core.telegram.org/api/end-to-end#sending-a-request.
func CheckDH(g int, p *big.Int) error {
	// The client is expected to check whether p is a safe 2048-bit prime
	// (meaning that both p and (p-1)/2 are prime, and that 2^2047 < p < 2^2048).
	// FIXME(tdakkota): we check that 2^2047 <= p < 2^2048
	// 	but docs says to check 2^2047 < p < 2^2048.
	//
	// TDLib check 2^2047 <= too:
	// https://github.com/tdlib/td/blob/d161323858a782bc500d188b9ae916982526c262/td/mtproto/DhHandshake.cpp#L23
	if p.BitLen() != RSAKeyBits {
		return errors.New("p should be 2^2047 < p < 2^2048")
	}

	if err := CheckGP(g, p); err != nil {
		return err
	}

	return checkPrime(p)
}

func checkPrime(p *big.Int) error {
	if !Prime(p) {
		return errors.New("p is not prime number")
	}

	sub := big.NewInt(0).Sub(p, big.NewInt(1))
	pr := sub.Quo(sub, big.NewInt(2))
	if !Prime(pr) {
		return errors.New("(p-1)/2 is not prime number")
	}

	return nil
}
