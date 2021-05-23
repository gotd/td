package telegram

import (
	"context"

	"github.com/gotd/td/tg"
)

// AuthBot performs bot authentication request.
//
// Deprecated: use auth.Client.
func (c *Client) AuthBot(ctx context.Context, token string) (*tg.AuthAuthorization, error) {
	return c.Auth().Bot(ctx, token)
}
