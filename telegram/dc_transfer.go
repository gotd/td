package telegram

import (
	"context"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
)

func (c *Client) dcTransferSetup(dcID int) manager.SetupCallback {
	return func(ctx context.Context, invoker tg.Invoker) error {
		// Run export/import authorization only when the connection is already up.
		_, err := c.transfer(ctx, tg.NewClient(invoker), dcID)
		if auth.IsUnauthorized(err) {
			return nil
		}
		return err
	}
}
