package telegram

import (
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/qrlogin"
)

// Auth returns auth client.
func (c *Client) Auth() *auth.Client {
	return auth.NewClient(
		c.tg, c.rand, c.appID, c.appHash,
	)
}

// QR returns QR login helper.
func (c *Client) QR() qrlogin.QR {
	return qrlogin.NewQR(
		c.tg,
		c.appID,
		c.appHash,
		qrlogin.Options{Migrate: c.MigrateTo},
	)
}
