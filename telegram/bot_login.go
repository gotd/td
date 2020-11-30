package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/ernado/td/tg"
)

// BotLogin performs bot authorization request.
func (c *Client) BotLogin(ctx context.Context, token string) error {
	var res tg.AuthAuthorizationBox
	if err := c.rpcContent(ctx, &tg.AuthImportBotAuthorizationRequest{
		APIID:        c.appID,
		APIHash:      c.appHash,
		BotAuthToken: token,
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
