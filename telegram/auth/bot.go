package auth

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// Bot performs bot authentication request.
func (c *Client) Bot(ctx context.Context, token string) (*tg.AuthAuthorization, error) {
	auth, err := c.api.AuthImportBotAuthorization(ctx, &tg.AuthImportBotAuthorizationRequest{
		APIID:        c.appID,
		APIHash:      c.appHash,
		BotAuthToken: token,
	})
	if err != nil {
		return nil, err
	}
	result, err := checkResult(auth)
	if err != nil {
		return nil, errors.Wrap(err, "check")
	}
	return result, nil
}
