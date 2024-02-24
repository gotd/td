package srp

import (
	"io"
	"math/big"

	"github.com/go-faster/errors"
)

// computeXV computes following numbers
//
// `x = PH2(password, salt1, salt2)`
// `v = pow(g, x) mod p`
//
// TDLib uses terms `clientSalt` for `salt1` and `serverSalt` for `salt2`.
func (s SRP) computeXV(password, clientSalt, serverSalt []byte, g, p *big.Int) (x, v *big.Int) {
	// `x = PH2(password, salt1, salt2)`
	x = new(big.Int).SetBytes(s.secondary(password, clientSalt, serverSalt))
	// `v = pow(g, x) mod p`
	v = new(big.Int).Exp(g, x, p)
	return x, v
}

// NewHash computes new user password hash using parameters from server.
//
// See https://core.telegram.org/api/srp#setting-a-new-2fa-password.
//
// TDLib implementation:
// See https://github.com/tdlib/td/blob/fa8feefed70d64271945e9d5fd010b957d93c8cd/td/telegram/PasswordManager.cpp#L57.
//
// TDesktop implementation:
// See https://github.com/telegramdesktop/tdesktop/blob/v3.4.8/Telegram/SourceFiles/core/core_cloud_password.cpp#L68.
func (s SRP) NewHash(password []byte, i Input) (hash, newSalt []byte, _ error) {
	// Generate a new new_password_hash using the KDF algorithm specified in the new_settings,
	// just append 32 sufficiently random bytes to the salt1, first. Proceed as for checking passwords with SRP,
	// just stop at the generation of the v parameter, and use it as new_password_hash:
	p := new(big.Int).SetBytes(i.P)
	if err := checkInput(i.G, p); err != nil {
		return nil, nil, errors.Wrap(err, "validate algo")
	}

	// Make a copy.
	newClientSalt := append([]byte(nil), i.Salt1...)
	newClientSalt = append(newClientSalt, make([]byte, 32)...)
	// ... append 32 sufficiently random bytes to the salt1 ...
	if _, err := io.ReadFull(s.random, newClientSalt[len(newClientSalt)-32:]); err != nil {
		return nil, nil, err
	}

	_, v := s.computeXV(password, newClientSalt, i.Salt2, big.NewInt(int64(i.G)), p)
	// As usual in big endian form, padded to 2048 bits.
	padded, _ := s.pad256FromBig(v)
	return padded[:], newClientSalt, nil
}
