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
	users, err := c.rpc().UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})
	if err != nil {
		return nil, err
	}

	user, ok := tg.UserClassArray(users).FirstAsNotEmpty()
	if !ok {
		return nil, xerrors.Errorf("users response count: %v", users)
	}

	return user, nil
}
