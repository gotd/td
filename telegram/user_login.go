package telegram

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto/srp"
	"github.com/gotd/td/tg"
)

func (c *Client) passwordLogin(ctx context.Context, auth UserAuthenticator) (tg.AuthAuthorizationClass, error) {
	password, err := auth.Password(ctx)
	if err != nil {
		return nil, xerrors.Errorf("failed to get password: %w", err)
	}

	p, err := c.tg.AccountGetPassword(ctx)
	if err != nil {
		return nil, xerrors.Errorf("failed to get SRP parameters: %w", err)
	}

	algo, ok := p.CurrentAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
	if !ok {
		return nil, xerrors.Errorf("unsupported algo: %T", p.CurrentAlgo)
	}

	s := srp.NewSRP(c.rand)
	a, err := s.Hash([]byte(password), p.SrpB, p.SecureRandom, srp.Input(*algo))
	if err != nil {
		return nil, xerrors.Errorf("failed to create SRP answer: %w", err)
	}

	return c.tg.AuthCheckPassword(ctx, &tg.InputCheckPasswordSRP{
		SrpID: p.SrpID,
		A:     a.A,
		M1:    a.M1,
	})
}

var errPasswordAuthNeed = errors.New("need password auth")

func (c *Client) codeLogin(ctx context.Context, auth UserAuthenticator, settings tg.CodeSettings) (tg.AuthAuthorizationClass, error) {
	phone, err := auth.Phone(ctx)
	if err != nil {
		return nil, xerrors.Errorf("failed to get phone: %w", err)
	}

	sentCode, err := c.tg.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
		PhoneNumber: phone,
		APIID:       c.appID,
		APIHash:     c.appHash,
		Settings:    settings,
	})
	if err != nil {
		return nil, xerrors.Errorf("failed to send code: %w", err)
	}

	code, err := auth.Code(ctx)
	if err != nil {
		return nil, xerrors.Errorf("failed to get code: %w", err)
	}

	a, err := c.tg.AuthSignIn(ctx, &tg.AuthSignInRequest{
		PhoneNumber:   phone,
		PhoneCodeHash: sentCode.PhoneCodeHash,
		PhoneCode:     code,
	})
	if err != nil {
		// TODO:(tdakkota) find better way to check it
		if strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") {
			return nil, errPasswordAuthNeed
		}
		return nil, xerrors.Errorf("failed to send code: %w", err)
	}

	return a, nil
}

func (c *Client) userLogin(ctx context.Context, auth UserAuthenticator, settings tg.CodeSettings) (tg.AuthAuthorizationClass, error) {
	a, err := c.codeLogin(ctx, auth, settings)
	switch {
	case err == nil:
		return a, nil
	case errors.Is(err, errPasswordAuthNeed):
		return c.passwordLogin(ctx, auth)
	default:
		return nil, err
	}
}

// UserLogin performs user authorization request.
func (c *Client) UserLogin(ctx context.Context, auth UserAuthenticator, settings tg.CodeSettings) error {
	a, err := c.userLogin(ctx, auth, settings)
	if err != nil {
		return err
	}

	switch a.(type) {
	case *tg.AuthAuthorization:
		// Ok.
		return nil
	default:
		return xerrors.Errorf("got unexpected response %T", auth)
	}
}
