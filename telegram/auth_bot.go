package telegram

import (
	"context"

	"github.com/gotd/td/tg"
)

// AuthBot performs bot authentication request.
func (c *Client) AuthBot(ctx context.Context, token string) (*tg.User, error) {
	auth, err := c.tg.AuthImportBotAuthorization(ctx, &tg.AuthImportBotAuthorizationRequest{
		APIID:        c.appID,
		APIHash:      c.appHash,
		BotAuthToken: token,
	})
	if err != nil {
		return nil, err
	}
	return checkAuthResult(auth)
}
