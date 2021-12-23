package peers

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

// getUserFull gets tg.UserFull using given tg.InputUserClass.
func (m *Manager) getUserFull(ctx context.Context, p tg.InputUserClass) (*tg.UserFull, error) {
	if userID, ok := m.getIDFromInputUser(p); ok && !m.needsUpdateFull(userPeerID(userID)) {
		// TODO(tdakkota): save full self.
		u, found, err := m.cache.FindUserFull(ctx, userID)
		if err == nil && found {
			u.SetFlags()
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

// getChatFull gets tg.ChatFull using given id.
func (m *Manager) getChatFull(ctx context.Context, p int64) (*tg.ChatFull, error) {
	if !m.needsUpdateFull(chatPeerID(p)) {
		c, found, err := m.cache.FindChatFull(ctx, p)
		if err == nil && found {
			c.SetFlags()
			return c, nil
		}
		if err != nil {
			m.logger.Warn("Find full chat error", zap.Int64("chat_id", p), zap.Error(err))
		}
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

// getChannelFull gets tg.ChannelFull using given tg.InputChannelClass.
func (m *Manager) getChannelFull(ctx context.Context, p tg.InputChannelClass) (*tg.ChannelFull, error) {
	if id, ok := getIDFromInputChannel(p); ok && !m.needsUpdateFull(channelPeerID(id)) {
		c, found, err := m.cache.FindChannelFull(ctx, id)
		if err == nil && found {
			c.SetFlags()
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
