package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/tg"
)

func (c *Client) PasswordLogin(ctx context.Context, password string) (tg.AuthAuthorizationClass, error) {
	p, err := c.tg.AccountGetPassword(ctx)
	if err != nil {
		return nil, xerrors.Errorf("failed to get SRP parameters: %w", err)
	}

	algo, ok := p.CurrentAlgo.(*tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow)
	if !ok {
		return nil, xerrors.Errorf("unsupported algo: %T", p.CurrentAlgo)
	}

	srp := crypto.NewSRP(c.rand)
	a, err := srp.Auth([]byte(password), p.SrpB, p.SecureRandom, crypto.SRPInput(*algo))
	if err != nil {
		return nil, xerrors.Errorf("failed to create SRP answer: %w", err)
	}

	return c.tg.AuthCheckPassword(ctx, &tg.InputCheckPasswordSRP{
		SrpID: p.SrpID,
		A:     a.A,
		M1:    a.M1,
	})
}
