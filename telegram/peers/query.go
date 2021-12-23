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
		if ok && !m.needsUpdate(userPeerID(u.ID)) {
			return u, nil
		}
	default:
		userID, ok := m.getIDFromInputUser(p)
		if !ok || m.needsUpdate(userPeerID(userID)) {
			break
		}

		if me, ok := m.me.Load(); ok && me.GetID() == userID {
			return me, nil
		}

		u, found, err := m.cache.FindUser(ctx, userID)
		if err == nil && found {
			u.SetFlags()
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

	return user, nil
}

// getChat gets tg.Chat using given id.
func (m *Manager) getChat(ctx context.Context, p int64) (*tg.Chat, error) {
	if !m.needsUpdate(chatPeerID(p)) {
		c, found, err := m.cache.FindChat(ctx, p)
		if err == nil && found {
			c.SetFlags()
			return c, nil
		}
		if err != nil {
			m.logger.Warn("Find chat error", zap.Int64("chat_id", p), zap.Error(err))
		}
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

	var found tg.ChatClass
	for _, chat := range chats {
		switch chat := chat.(type) {
		case *tg.Chat:
			if chat.ID == id {
				found = chat
				break
			}
		}
	}

	ch, ok := found.(*tg.Chat)
	if !ok {
		// TODO(tdakkota): get better error for forbidden.
		return nil, errors.Errorf("got unexpected type %T", found)
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
	if id, ok := getIDFromInputChannel(p); ok && !m.needsUpdate(channelPeerID(id)) {
		c, found, err := m.cache.FindChannel(ctx, id)
		if err == nil && found {
			c.SetFlags()
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

	var found tg.ChatClass
	if inputHasID, ok := p.AsNotEmpty(); ok {
		id := inputHasID.GetChannelID()
		for _, chat := range chats {
			switch chat := chat.(type) {
			case *tg.Channel:
				if chat.ID == id {
					found = chat
					break
				}
			}
		}
	} else {
		found = chats[0]
	}

	ch, ok := found.(*tg.Channel)
	if !ok {
		// TODO(tdakkota): get better error for forbidden.
		return nil, errors.Errorf("got unexpected type %T", found)
	}

	return ch, nil
}
