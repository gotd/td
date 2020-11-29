package telegram

import (
	"context"
	"errors"

	"golang.org/x/xerrors"

	"github.com/ernado/td/tg"
)

// AuthStatus represents authorization status.
type AuthStatus struct {
	// Authorized is true if client is authorized.
	Authorized bool
}

// AuthStatus gets authorization status of client.
func (c *Client) AuthStatus(ctx context.Context) (*AuthStatus, error) {
	var res tg.UpdatesState
	if err := c.rpcNoAck(ctx, &tg.UpdatesGetStateRequest{}, &res); err != nil {
		var rpcErr *Error
		if errors.As(err, &rpcErr) {
			// Not authorized.
			// TODO(ernado): Check for specific code.
			return &AuthStatus{
				Authorized: false,
			}, nil
		}
		return nil, xerrors.Errorf("failed to perform request: %w", err)
	}
	return &AuthStatus{Authorized: true}, nil
}
