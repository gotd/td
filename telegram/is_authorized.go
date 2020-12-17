package telegram

import (
	"context"
	"errors"

	"github.com/gotd/td/tg"
)

// AuthStatus represents authorization status.
type AuthStatus struct {
	// Authorized is true if client is authorized.
	Authorized bool
	// User is current User object.
	User *tg.User
}

// AuthStatus gets authorization status of client.
func (c *Client) AuthStatus(ctx context.Context) (*AuthStatus, error) {
	u, err := c.Self(ctx)
	if err != nil {
		var rpcErr *Error
		if errors.As(err, &rpcErr) && rpcErr.Message == "AUTH_KEY_UNREGISTERED" {
			return &AuthStatus{}, nil
		}

		return nil, err
	}

	return &AuthStatus{
		Authorized: true,
		User:       u,
	}, nil
}
