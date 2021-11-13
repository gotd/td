package crypto

import (
	"math/big"

	"github.com/go-faster/errors"
)

// CheckGP checks whether g generates a cyclic subgroup of prime order (p-1)/2, i.e. is a quadratic residue mod p.
// Also check that g is 2, 3, 4, 5, 6 or 7.
//
// This function is needed by some Telegram algorithms(Key generation, SRP 2FA).
//
// See https://core.telegram.org/mtproto/auth_key.
//
// See https://core.telegram.org/api/srp.
func CheckGP(g int, p *big.Int) error {
	// Since g is always equal to 2, 3, 4, 5, 6 or 7,
	// this is easily done using quadratic reciprocity law, yielding a simple condition on p mod 4g -- namely,
	var result bool
	switch g {
	case 2:
		// p mod 8 = 7 for g = 2;
		result = checkSubgroup(p, 8, 7)
	case 3:
		// p mod 3 = 2 for g = 3;
		result = checkSubgroup(p, 3, 2)
	case 4:
		// no extra condition for g = 4
		result = true
	case 5:
		// p mod 5 = 1 or 4 for g = 5;
		result = checkSubgroup(p, 5, 1, 4)
	case 6:
		// p mod 24 = 19 or 23 for g = 6;
		result = checkSubgroup(p, 24, 19, 23)
	case 7:
		// and p mod 7 = 3, 5 or 6 for g = 7.
		result = checkSubgroup(p, 7, 3, 5, 6)
	default:
		return errors.Errorf("unexpected g = %d: g should be equal to 2, 3, 4, 5, 6 or 7", g)
	}

	if !result {
		return errors.New("g should be a quadratic residue mod p")
	}

	return nil
}

func checkSubgroup(p *big.Int, divider int64, expected ...int64) bool {
	rem := new(big.Int).Rem(p, big.NewInt(divider)).Int64()

	for _, e := range expected {
		if rem == e {
			return true
		}
	}

	return false
}
