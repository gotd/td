package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/crypto/srp"
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

// UpdatePasswordOptions is options structure for UpdatePassword.
type UpdatePasswordOptions struct {
	// Hint is new password hint.
	Hint string
	// Password is password callback.
	//
	// If password was requested and Password is nil, ErrPasswordNotProvided error will be returned.
	Password func(ctx context.Context) (string, error)
}

// UpdatePassword sets new cloud password for this account.
//
// See https://core.telegram.org/api/srp#setting-a-new-2fa-password.
func (c *Client) UpdatePassword(
	ctx context.Context,
	newPassword string,
	opts UpdatePasswordOptions,
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
		if opts.Password == nil {
			return ErrPasswordNotProvided
		}

		oldPassword, err := opts.Password(ctx)
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
			Hint:            opts.Hint,
		},
	}); err != nil {
		return errors.Wrap(err, "update password")
	}
	return nil
}

// ResetFailedWaitError reports that you recently requested a password reset that was cancel and need to wait until the
// specified date before requesting another reset.
type ResetFailedWaitError struct {
	Result tg.AccountResetPasswordFailedWait
}

// Until returns time required to wait.
func (r ResetFailedWaitError) Until() time.Duration {
	retryDate := time.Unix(int64(r.Result.RetryDate), 0)
	return time.Until(retryDate)
}

// Error implements error.
func (r *ResetFailedWaitError) Error() string {
	return fmt.Sprintf("wait to reset password (%s)", r.Until())
}

// ResetPassword resets cloud password and returns time to wait until reset be performed.
// If time is zero, password was successfully reset.
//
// May return ResetFailedWaitError.
//
// See https://core.telegram.org/api/srp#password-reset.
func (c *Client) ResetPassword(ctx context.Context) (time.Time, error) {
	r, err := c.api.AccountResetPassword(ctx)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "reset password")
	}
	switch v := r.(type) {
	case *tg.AccountResetPasswordFailedWait:
		return time.Time{}, &ResetFailedWaitError{Result: *v}
	case *tg.AccountResetPasswordRequestedWait:
		return time.Unix(int64(v.UntilDate), 0), nil
	case *tg.AccountResetPasswordOk:
		return time.Time{}, nil
	default:
		return time.Time{}, errors.Errorf("unexpected type %T", v)
	}
}

// CancelPasswordReset cancels password reset.
//
// See https://core.telegram.org/api/srp#password-reset.
func (c *Client) CancelPasswordReset(ctx context.Context) error {
	if _, err := c.api.AccountDeclinePasswordReset(ctx); err != nil {
		return errors.Wrap(err, "cancel password reset")
	}
	return nil
}
