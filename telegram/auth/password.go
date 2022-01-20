package auth

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/crypto/srp"
	"github.com/gotd/td/tg"
)

// PasswordHash computes password hash to log in.
//
// See https://core.telegram.org/api/srp#checking-the-password-with-srp.
func PasswordHash(
	password []byte,
	srpID int64,
	srpB, secureRandom []byte,
	alg tg.PasswordKdfAlgoClass,
) (*tg.InputCheckPasswordSRP, error) {
	s := srp.NewSRP(crypto.DefaultRand())

	algo, ok := alg.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
	if !ok {
		return nil, errors.Errorf("unsupported algo: %T", alg)
	}

	a, err := s.Hash(password, srpB, secureRandom, srp.Input(*algo))
	if err != nil {
		return nil, errors.Wrap(err, "create SRP answer")
	}

	return &tg.InputCheckPasswordSRP{
		SRPID: srpID,
		A:     a.A,
		M1:    a.M1,
	}, nil
}

// NewPasswordHash computes new password hash to update password.
//
// Notice that NewPasswordHash mutates given alg.
//
// See https://core.telegram.org/api/srp#setting-a-new-2fa-password.
func NewPasswordHash(
	password []byte,
	algo *tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow,
) (hash []byte, _ error) {
	s := srp.NewSRP(crypto.DefaultRand())

	hash, newSalt, err := s.NewHash(password, srp.Input(*algo))
	if err != nil {
		return nil, errors.Wrap(err, "create SRP answer")
	}
	algo.Salt1 = newSalt

	return hash, nil
}

var (
	emptyPassword tg.InputCheckPasswordSRPClass = &tg.InputCheckPasswordEmpty{}
)

// UpdatePassword sets new password for this account.
func (c *Client) UpdatePassword(
	ctx context.Context,
	hint, newPassword string,
	pass func(ctx context.Context) (string, error),
) error {
	p, err := c.api.AccountGetPassword(ctx)
	if err != nil {
		return errors.Wrap(err, "get SRP parameters")
	}

	algo, ok := p.NewAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
	if !ok {
		return errors.Errorf("unsupported algo: %T", p.NewAlgo)
	}

	newHash, err := NewPasswordHash([]byte(newPassword), algo)
	if err != nil {
		return errors.Wrap(err, "compute new password hash")
	}

	var old = emptyPassword
	if p.HasPassword {
		if pass == nil {
			return ErrPasswordNotProvided
		}

		oldPassword, err := pass(ctx)
		if err != nil {
			return errors.Wrap(err, "get password")
		}

		hash, err := PasswordHash([]byte(oldPassword), p.SRPID, p.SRPB, p.SecureRandom, p.CurrentAlgo)
		if err != nil {
			return errors.Wrap(err, "compute old password hash")
		}
		old = hash
	}

	if _, err := c.api.AccountUpdatePasswordSettings(ctx, &tg.AccountUpdatePasswordSettingsRequest{
		Password: old,
		NewSettings: tg.AccountPasswordInputSettings{
			NewAlgo:         algo,
			NewPasswordHash: newHash,
			Hint:            hint,
		},
	}); err != nil {
		return errors.Wrap(err, "update password")
	}
	return nil
}
