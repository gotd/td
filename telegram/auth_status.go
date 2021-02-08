package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tg"
)

// AuthStatus represents authorization status.
type AuthStatus struct {
	// Authorized is true if client is authorized.
	Authorized bool
	// User is current User object.
	User *tg.User
}

func unauthorized(err error) bool {
	return mtproto.IsErr(err, "AUTH_KEY_UNREGISTERED")
}

// AuthStatus gets authorization status of client.
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
func (c *Client) AuthIfNecessary(ctx context.Context, flow AuthFlow) error {
	auth, err := c.AuthStatus(ctx)
	if err != nil {
		return xerrors.Errorf("get auth status: %w", err)
	}

	if !auth.Authorized {
		if err := flow.Run(ctx, c); err != nil {
			return xerrors.Errorf("auth flow: %w", err)
		}
	}

	return nil
}
