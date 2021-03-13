package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// AuthBot performs bot authentication request.
func (c *Client) AuthBot(ctx context.Context, token string) error {
	auth, err := c.tg.AuthImportBotAuthorization(ctx, &tg.AuthImportBotAuthorizationRequest{
		APIID:        c.appID,
		APIHash:      c.appHash,
		BotAuthToken: token,
	})
	if err != nil {
		return err
	}
	if err := checkAuthResult(auth); err != nil {
		return xerrors.Errorf("check: %w", err)
	}
	return nil
}
