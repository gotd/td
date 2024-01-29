package telegram

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

func (c *Client) exportAuth(ctx context.Context, dcID int) (*tg.AuthExportedAuthorization, error) {
	export, err := c.tg.AuthExportAuthorization(ctx, dcID)
	if err != nil {
		return nil, errors.Wrapf(err, "export auth to %d", dcID)
	}

	return export, nil
}

// AuthTransferHandler is a function that is called during authorization transfer.
//
// The fn callback should be serialized by user id via external locking.
// You can call [Client.Self] to acquire current user id.
//
// The fn callback must return fn error if any.
type AuthTransferHandler func(ctx context.Context, client *Client, fn func(context.Context) error) error

func noopOnTransfer(ctx context.Context, _ *Client, fn func(context.Context) error) error {
	return fn(ctx)
}

// transfer exports current authorization and imports it to another DC.
// See https://core.telegram.org/api/datacenter#authorization-transfer.
func (c *Client) transfer(ctx context.Context, to *tg.Client, dc int) (tg.AuthAuthorizationClass, error) {
	var out tg.AuthAuthorizationClass
	if err := c.onTransfer(ctx, c, func(ctx context.Context) error {
		auth, err := c.exportAuth(ctx, dc)
		if err != nil {
			return errors.Wrapf(err, "export to %d", dc)
		}

		req := &tg.AuthImportAuthorizationRequest{}
		req.FillFrom(auth)
		r, err := to.AuthImportAuthorization(ctx, req)
		if err != nil {
			return errors.Wrapf(err, "import from %d", dc)
		}

		out = r
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "onTransfer")
	}
	return out, nil
}
