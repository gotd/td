package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/tg"
)

// BotLogin wraps credentials that are required to login as bot.
type BotLogin struct {
	// ID is api_id.
	ID int
	// Hash is api_hash.
	Hash string
	// Token is bot auth token.
	Token string
}

type authBox struct {
	Auth tg.AuthAuthorizationClass
}

func (a *authBox) Decode(b *bin.Buffer) error {
	v, err := tg.DecodeAuthAuthorization(b)
	if err != nil {
		return err
	}
	a.Auth = v
	return nil
}

// BotLogin performs bot authorization request.
func (c *Client) BotLogin(ctx context.Context, login BotLogin) error {
	var res authBox
	if err := c.do(ctx, &tg.AuthImportBotAuthorizationRequest{
		APIID:        login.ID,
		APIHash:      login.Hash,
		BotAuthToken: login.Token,
	}, &res); err != nil {
		return xerrors.Errorf("failed to do request: %w", err)
	}
	switch res.Auth.(type) {
	case *tg.AuthAuthorization:
		// Ok.
		return nil
	default:
		return xerrors.Errorf("got unexpected response %T", res.Auth)
	}
}
