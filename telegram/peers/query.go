package peers

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

func (m *Manager) getIDFromInputUser(p tg.InputUserClass) (int64, bool) {
	switch p := p.(type) {
	case *tg.InputUserSelf:
		return m.myID()
	case *tg.InputUser:
		return p.UserID, true
	case *tg.InputUserFromMessage:
		return p.UserID, true
	default:
		return 0, false
	}
}

// getUser gets tg.User using given tg.InputUserClass.
func (m *Manager) getUser(ctx context.Context, p tg.InputUserClass) (*tg.User, error) {
	switch p := p.(type) {
	case *tg.InputUserSelf:
		u, ok := m.me.Load()
		if ok {
			return u, nil
		}
	default:
		userID, ok := m.getIDFromInputUser(p)
		if !ok {
			break
		}

		u, found, err := m.cache.FindUser(ctx, userID)
		if err == nil && found {
			return u, nil
		}
		if err != nil {
			m.logger.Warn("Find user error", zap.Int64("user_id", userID), zap.Error(err))
		}
	}

	return m.updateUser(ctx, p)
}

// updateUser forcibly updates tg.User using given tg.InputUserClass.
func (m *Manager) updateUser(ctx context.Context, p tg.InputUserClass) (*tg.User, error) {
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
	if user.Self {
		m.me.Store(user)
	}
	if user.Support {
		m.support.Store(user)
	}

	return user, nil
}

// getUserFull gets tg.UserFull using given tg.InputUserClass.
func (m *Manager) getUserFull(ctx context.Context, p tg.InputUserClass) (*tg.UserFull, error) {
	userID, ok := m.getIDFromInputUser(p)
	if ok {
		u, found, err := m.cache.FindUserFull(ctx, userID)
		if err == nil && found {
			return u, nil
		}
		if err != nil {
			m.logger.Warn("Find full user error", zap.Int64("user_id", userID), zap.Error(err))
		}
	}
	return m.updateUserFull(ctx, p)
}

// updateUserFull forcibly updates tg.UserFull using given tg.InputUserClass.
func (m *Manager) updateUserFull(ctx context.Context, p tg.InputUserClass) (*tg.UserFull, error) {
	r, err := m.api.UsersGetFullUser(ctx, p)
	if err != nil {
		return nil, errors.Wrap(err, "get full user")
	}

	if err := m.applyEntities(ctx, r.GetUsers(), r.GetChats()); err != nil {
		return nil, err
	}

	if err := m.applyFullUser(ctx, &r.FullUser); err != nil {
		return nil, errors.Wrap(err, "update full user")
	}

	cp := r.FullUser
	return &cp, nil
}

// getChat gets tg.Chat using given id.
func (m *Manager) getChat(ctx context.Context, p int64) (*tg.Chat, error) {
	c, found, err := m.cache.FindChat(ctx, p)
	if err == nil && found {
		return c, nil
	}
	if err != nil {
		m.logger.Warn("Find chat error", zap.Int64("chat_id", p), zap.Error(err))
	}
	return m.updateChat(ctx, p)
}

// updateChat forcibly updates tg.Chat using given id.
func (m *Manager) updateChat(ctx context.Context, id int64) (*tg.Chat, error) {
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

// getChatFull gets tg.ChatFull using given id.
func (m *Manager) getChatFull(ctx context.Context, p int64) (*tg.ChatFull, error) {
	c, found, err := m.cache.FindChatFull(ctx, p)
	if err == nil && found {
		return c, nil
	}
	if err != nil {
		m.logger.Warn("Find full chat error", zap.Int64("chat_id", p), zap.Error(err))
	}
	return m.updateChatFull(ctx, p)
}

// updateChatFull forcibly updates tg.ChatFull using given id.
func (m *Manager) updateChatFull(ctx context.Context, id int64) (*tg.ChatFull, error) {
	r, err := m.api.MessagesGetFullChat(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "get full chat")
	}

	if err := m.applyEntities(ctx, r.GetUsers(), r.GetChats()); err != nil {
		return nil, err
	}

	ch, ok := r.FullChat.(*tg.ChatFull)
	if !ok {
		return nil, errors.Errorf("got unexpected type %T", r.FullChat)
	}

	if err := m.applyFullChat(ctx, ch); err != nil {
		return nil, errors.Wrap(err, "update full chat")
	}

	return ch, nil
}

func getIDFromInputChannel(p tg.InputChannelClass) (int64, bool) {
	switch p := p.(type) {
	case *tg.InputChannel:
		return p.ChannelID, true
	case *tg.InputChannelFromMessage:
		return p.ChannelID, true
	default:
		return 0, false
	}
}

// getChannel gets tg.Channel using given tg.InputChannelClass.
func (m *Manager) getChannel(ctx context.Context, p tg.InputChannelClass) (*tg.Channel, error) {
	if id, ok := getIDFromInputChannel(p); ok {
		c, found, err := m.cache.FindChannel(ctx, id)
		if err == nil && found {
			return c, nil
		}
		if err != nil {
			m.logger.Warn("Find channel error", zap.Int64("channel_id", id), zap.Error(err))
		}
	}
	return m.updateChannel(ctx, p)
}

// updateChannel forcibly updates tg.Channel using given tg.InputChannelClass.
func (m *Manager) updateChannel(ctx context.Context, p tg.InputChannelClass) (*tg.Channel, error) {
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

// getChannelFull gets tg.ChannelFull using given tg.InputChannelClass.
func (m *Manager) getChannelFull(ctx context.Context, p tg.InputChannelClass) (*tg.ChannelFull, error) {
	if id, ok := getIDFromInputChannel(p); ok {
		c, found, err := m.cache.FindChannelFull(ctx, id)
		if err == nil && found {
			return c, nil
		}
		if err != nil {
			m.logger.Warn("Find channel error", zap.Int64("channel_id", id), zap.Error(err))
		}
	}
	return m.updateChannelFull(ctx, p)
}

// updateChannelFull forcibly updates tg.ChannelFull using given tg.InputChannelClass.
func (m *Manager) updateChannelFull(ctx context.Context, p tg.InputChannelClass) (*tg.ChannelFull, error) {
	r, err := m.api.ChannelsGetFullChannel(ctx, p)
	if err != nil {
		return nil, errors.Wrap(err, "get full channel")
	}

	if err := m.applyEntities(ctx, r.GetUsers(), r.GetChats()); err != nil {
		return nil, err
	}

	ch, ok := r.FullChat.(*tg.ChannelFull)
	if !ok {
		return nil, errors.Errorf("got unexpected type %T", r.FullChat)
	}

	if err := m.applyFullChannel(ctx, ch); err != nil {
		return nil, errors.Wrap(err, "update full channel")
	}

	return ch, nil
}
