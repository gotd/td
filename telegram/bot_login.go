package telegram

import (
	"context"

	"golang.org/x/xerrors"

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

// BotLogin performs bot authorization request.
func (c *Client) BotLogin(ctx context.Context, login BotLogin) error {
	var res tg.AuthAuthorizationBox
	if err := c.do(ctx, &tg.AuthImportBotAuthorizationRequest{
		APIID:        login.ID,
		APIHash:      login.Hash,
		BotAuthToken: login.Token,
	}, &res); err != nil {
		return xerrors.Errorf("failed to do request: %w", err)
	}
	switch res.AuthAuthorization.(type) {
	case *tg.AuthAuthorization:
		// Ok.
		return nil
	default:
		return xerrors.Errorf("got unexpected response %T", res.AuthAuthorization)
	}
}
