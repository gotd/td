package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// Self returns current user.
//
// You can use tg.User.Bot to check whether current user is bot.
func (c *Client) Self(ctx context.Context) (*tg.User, error) {
	users, err := c.RPC.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})
	if err != nil {
		return nil, err
	}

	if len(users) != 1 {
		return nil, xerrors.Errorf("bad users count: %d", len(users))
	}

	user, ok := users[0].(*tg.User)
	if !ok {
		return nil, xerrors.Errorf("unexpected user type: %T", users[0])
	}

	return user, nil
}
