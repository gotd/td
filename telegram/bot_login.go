package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// BotLogin performs bot authorization request.
func (c *Client) BotLogin(ctx context.Context, token string) error {
	auth, err := c.tg.AuthImportBotAuthorization(ctx, &tg.AuthImportBotAuthorizationRequest{
		APIID:        c.appID,
		APIHash:      c.appHash,
		BotAuthToken: token,
	})
	if err != nil {
		return err
	}
	switch auth.(type) {
	case *tg.AuthAuthorization:
		// Ok.
		return nil
	default:
		return xerrors.Errorf("got unexpected response %T", auth)
	}
}
