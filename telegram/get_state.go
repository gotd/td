package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/ernado/td/tg"
)

func (c *Client) GetState(ctx context.Context) (*tg.UpdatesState, error) {
	var res tg.UpdatesState
	if err := c.rpcNoAck(ctx, &tg.UpdatesGetStateRequest{}, &res); err != nil {
		return nil, xerrors.Errorf("failed to do: %w", err)
	}
	return &res, nil
}
