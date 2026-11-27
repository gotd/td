package crypto

import (
	"math/big"
	"sync"

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
//
// A p that passes is remembered, so checking it again is cheap.
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
	if isSafePrime(p) {
		return nil
	}

	if !Prime(p) {
		return errors.New("p is not prime number")
	}

	sub := big.NewInt(0).Sub(p, big.NewInt(1))
	pr := sub.Quo(sub, big.NewInt(2))
	if !Prime(pr) {
		return errors.New("(p-1)/2 is not prime number")
	}

	rememberSafePrime(p)
	return nil
}

// maxSafePrimes bounds safePrimes. Telegram uses one well-known dh_prime, so
// a few slots are plenty; once they are taken, other primes are still
// checked, just not remembered.
const maxSafePrimes = 8

// safePrimes holds the primes that passed checkPrime, so the same dh_prime is
// not tested again on every key exchange, SRP check and call. Only primes that
// passed are kept. TDLib does the same, see DhCache:
// https://github.com/tdlib/td/blob/master/td/telegram/DhCache.cpp.
var safePrimes struct {
	mux    sync.RWMutex
	primes []*big.Int
}

func isSafePrime(p *big.Int) bool {
	safePrimes.mux.RLock()
	defer safePrimes.mux.RUnlock()

	return hasSafePrime(p)
}

func rememberSafePrime(p *big.Int) {
	safePrimes.mux.Lock()
	defer safePrimes.mux.Unlock()

	if len(safePrimes.primes) < maxSafePrimes && !hasSafePrime(p) {
		safePrimes.primes = append(safePrimes.primes, new(big.Int).Set(p))
	}
}

// hasSafePrime must be called with safePrimes.mux held.
func hasSafePrime(p *big.Int) bool {
	for _, known := range safePrimes.primes {
		if known.Cmp(p) == 0 {
			return true
		}
	}
	return false
}
