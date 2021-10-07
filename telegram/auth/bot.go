package auth

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
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
		return nil, xerrors.Errorf("check: %w", err)
	}
	return result, nil
}
