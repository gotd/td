package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// getUser gets User using given tg.InputUserClass.
func (m *Manager) getUser(ctx context.Context, p tg.InputUserClass) (*tg.User, error) {
	// TODO(tdakkota): batch requests.
	users, err := m.api.UsersGetUsers(ctx, []tg.InputUserClass{p})
	if err != nil {
		return nil, errors.Wrap(err, "get users")
	}

	if len(users) < 1 {
		return nil, errors.Errorf("got empty result for %+v", p)
	}

	if err := m.applyUsers(ctx, users...); err != nil {
		return nil, errors.Wrap(err, "update users")
	}

	user, ok := users[0].AsNotEmpty()
	if !ok {
		return nil, errors.New("got empty user")
	}

	return user, nil
}

// getChat gets Chat using given id.
func (m *Manager) getChat(ctx context.Context, id int64) (*tg.Chat, error) {
	r, err := m.api.MessagesGetChats(ctx, []int64{id})
	if err != nil {
		return nil, errors.Wrap(err, "get chats")
	}
	chats := r.GetChats()

	if len(chats) < 1 {
		return nil, errors.Errorf("got empty result for chat %d", id)
	}

	if err := m.applyChats(ctx, chats...); err != nil {
		return nil, errors.Wrap(err, "update chats")
	}

	ch, ok := chats[0].(*tg.Chat)
	if !ok {
		// TODO(tdakkota): get better error for forbidden.
		return nil, errors.Errorf("got unexpected type %T", chats[0])
	}

	return ch, nil
}

// getChannel gets Channel using given tg.InputChannelClass.
func (m *Manager) getChannel(ctx context.Context, p tg.InputChannelClass) (*tg.Channel, error) {
	r, err := m.api.ChannelsGetChannels(ctx, []tg.InputChannelClass{p})
	if err != nil {
		return nil, errors.Wrap(err, "get channels")
	}
	chats := r.GetChats()

	if len(chats) < 1 {
		return nil, errors.Errorf("got empty result for %+v", p)
	}

	if err := m.applyChats(ctx, chats...); err != nil {
		return nil, errors.Wrap(err, "update chats")
	}

	ch, ok := chats[0].(*tg.Channel)
	if !ok {
		// TODO(tdakkota): get better error for forbidden.
		return nil, errors.Errorf("got unexpected type %T", chats[0])
	}

	return ch, nil
}
