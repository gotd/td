package telegram

import (
	"context"

	"golang.org/x/xerrors"
)

func (c *Client) migrateToDc(ctx context.Context, dcID int) error {
	// TODO(ernado): re-implement
	_ = ctx
	_ = dcID
	return xerrors.New("not implemented")
}
