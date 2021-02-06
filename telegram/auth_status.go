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
	var rpcErr *mtproto.Error
	return xerrors.As(err, &rpcErr) && rpcErr.Message == "AUTH_KEY_UNREGISTERED"
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
