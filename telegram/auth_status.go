package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// AuthStatus represents authorization status.
//
// Deprecated: use auth package.
type AuthStatus struct {
	// Authorized is true if client is authorized.
	Authorized bool
	// User is current User object.
	User *tg.User
}

func unauthorized(err error) bool {
	return tgerr.Is(err, "AUTH_KEY_UNREGISTERED")
}

// AuthStatus gets authorization status of client.
//
// Deprecated: use auth package.
func (c *Client) AuthStatus(ctx context.Context) (*AuthStatus, error) {
	u, err := c.Self(ctx)
	if err != nil {
		if unauthorized(err) {
			return &AuthStatus{}, nil
		}

		return nil, err
	}

	return &AuthStatus{
		Authorized: true,
		User:       u,
	}, nil
}

// AuthIfNecessary runs given auth flow if current session is not authorized.
//
// Deprecated: use auth package.
func (c *Client) AuthIfNecessary(ctx context.Context, flow AuthFlow) error {
	auth, err := c.AuthStatus(ctx)
	if err != nil {
		return xerrors.Errorf("get auth status: %w", err)
	}
	if auth.Authorized {
		return nil
	}
	if err := flow.Run(ctx, c.Auth()); err != nil {
		return xerrors.Errorf("auth flow: %w", err)
	}
	return nil
}
