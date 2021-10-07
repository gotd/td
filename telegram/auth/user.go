package auth

import (
	"context"
	"errors"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/crypto/srp"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
)

// ErrPasswordInvalid means that password provided to Password is invalid.
//
// Note that telegram does not trim whitespace characters by default, check
// that provided password is expected and clean whitespaces if needed.
// You can use strings.TrimSpace(password) for this.
var ErrPasswordInvalid = errors.New("invalid password")

// Password performs login via secure remote password (aka 2FA).
//
// Method can be called after SignIn to provide password if requested.
func (c *Client) Password(ctx context.Context, password string) (*tg.AuthAuthorization, error) {
	p, err := c.api.AccountGetPassword(ctx)
	if err != nil {
		return nil, xerrors.Errorf("get SRP parameters: %w", err)
	}

	algo, ok := p.CurrentAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
	if !ok {
		return nil, xerrors.Errorf("unsupported algo: %T", p.CurrentAlgo)
	}

	s := srp.NewSRP(c.rand)
	a, err := s.Hash([]byte(password), p.SRPB, p.SecureRandom, srp.Input(*algo))
	if err != nil {
		return nil, xerrors.Errorf("create SRP answer: %w", err)
	}

	auth, err := c.api.AuthCheckPassword(ctx, &tg.InputCheckPasswordSRP{
		SRPID: p.SRPID,
		A:     a.A,
		M1:    a.M1,
	})
	if tg.IsPasswordHashInvalid(err) {
		return nil, ErrPasswordInvalid
	}
	if err != nil {
		return nil, xerrors.Errorf("check password: %w", err)
	}
	result, err := checkResult(auth)
	if err != nil {
		return nil, xerrors.Errorf("check: %w", err)
	}
	return result, nil
}

// SendCodeOptions defines how to send auth code to user.
type SendCodeOptions struct {
	// AllowFlashCall allows phone verification via phone calls.
	AllowFlashCall bool
	// Pass true if the phone number is used on the current device.
	// Ignored if AllowFlashCall is not set.
	CurrentNumber bool
	// If a token that will be included in eventually sent SMSs is required:
	// required in newer versions of android, to use the android SMS receiver APIs.
	AllowAppHash bool
}

// SendCode requests code for provided phone number, returning code hash
// and error if any. Use AuthFlow to reduce boilerplate.
//
// This method should be called first in user authentication flow.
func (c *Client) SendCode(ctx context.Context, phone string, options SendCodeOptions) (*tg.AuthSentCode, error) {
	var settings tg.CodeSettings
	if options.AllowAppHash {
		settings.SetAllowAppHash(true)
	}
	if options.AllowFlashCall {
		settings.SetAllowFlashcall(true)
	}
	if options.CurrentNumber {
		settings.SetCurrentNumber(true)
	}

	sentCode, err := c.api.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
		PhoneNumber: phone,
		APIID:       c.appID,
		APIHash:     c.appHash,
		Settings:    settings,
	})
	if err != nil {
		return nil, xerrors.Errorf("send code: %w", err)
	}
	return sentCode, nil
}

// ErrPasswordAuthNeeded means that 2FA auth is required.
//
// Call Client.Password to provide 2FA password.
var ErrPasswordAuthNeeded = errors.New("2FA required")

// SignIn performs sign in with provided user phone, code and code hash.
//
// If ErrPasswordAuthNeeded is returned, call Password to provide 2FA
// password.
//
// To obtain codeHash, use SendCode.
func (c *Client) SignIn(ctx context.Context, phone, code, codeHash string) (*tg.AuthAuthorization, error) {
	auth, err := c.api.AuthSignIn(ctx, &tg.AuthSignInRequest{
		PhoneNumber:   phone,
		PhoneCodeHash: codeHash,
		PhoneCode:     code,
	})
	if tgerr.Is(err, "SESSION_PASSWORD_NEEDED") {
		return nil, ErrPasswordAuthNeeded
	}
	if err != nil {
		return nil, xerrors.Errorf("sign in: %w", err)
	}
	result, err := checkResult(auth)
	if err != nil {
		return nil, xerrors.Errorf("check: %w", err)
	}
	return result, nil
}

// AcceptTOS accepts version of Terms Of Service.
func (c *Client) AcceptTOS(ctx context.Context, id tg.DataJSON) error {
	_, err := c.api.HelpAcceptTermsOfService(ctx, id)
	return err
}

// SignUp wraps parameters for SignUp.
type SignUp struct {
	PhoneNumber   string
	PhoneCodeHash string
	FirstName     string
	LastName      string
}

// SignUp registers a validated phone number in the system.
//
// To obtain codeHash, use SendCode.
// Use AuthFlow helper to handle authentication flow.
func (c *Client) SignUp(ctx context.Context, s SignUp) (*tg.AuthAuthorization, error) {
	auth, err := c.api.AuthSignUp(ctx, &tg.AuthSignUpRequest{
		LastName:      s.LastName,
		PhoneCodeHash: s.PhoneCodeHash,
		PhoneNumber:   s.PhoneNumber,
		FirstName:     s.FirstName,
	})
	if err != nil {
		return nil, xerrors.Errorf("request: %w", err)
	}
	result, err := checkResult(auth)
	if err != nil {
		return nil, xerrors.Errorf("check: %w", err)
	}
	return result, nil
}
