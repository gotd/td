package peers

import (
	"context"

	"go.uber.org/multierr"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

// SetChannelAccessHash implements updates.ChannelAccessHasher.
func (m *Manager) SetChannelAccessHash(ctx context.Context, userID, channelID, accessHash int64) error {
	myID, ok := m.myID()
	if !ok || myID != userID {
		return nil
	}
	return m.storage.Save(context.TODO(), Key{
		Prefix: channelPrefix,
		ID:     channelID,
	}, Value{
		AccessHash: accessHash,
	})
}

// GetChannelAccessHash implements updates.ChannelAccessHasher.
func (m *Manager) GetChannelAccessHash(ctx context.Context, userID, channelID int64) (accessHash int64, found bool, err error) {
	myID, ok := m.myID()
	if !ok || myID != userID {
		return 0, false, nil
	}
	v, found, err := m.storage.Find(context.TODO(), Key{
		Prefix: channelPrefix,
		ID:     channelID,
	})
	return v.AccessHash, found, err
}

// UpdateHook returns update middleware hook for collecting entities.
func (m *Manager) UpdateHook(next telegram.UpdateHandler) telegram.UpdateHandler {
	f := func(ctx context.Context, u tg.UpdatesClass) error {
		var (
			users   []tg.UserClass
			chats   []tg.ChatClass
			updates []tg.UpdateClass
		)
		switch u := u.(type) {
		case *tg.UpdatesCombined:
			users = u.GetUsers()
			chats = u.GetChats()
			updates = u.GetUpdates()
		case *tg.Updates:
			users = u.GetUsers()
			chats = u.GetChats()
			updates = u.GetUpdates()
		}
		m.applyUpdates(updates)
		applyErr := m.applyEntities(ctx, users, chats)
		handleErr := next.Handle(ctx, u)
		return multierr.Append(handleErr, applyErr)
	}
	return telegram.UpdateHandlerFunc(f)
}

// UserResolveHook creates entity.UserResolver attached to this Manager.
func (m *Manager) UserResolveHook(ctx context.Context) entity.UserResolver {
	return func(id int64) (tg.InputUserClass, error) {
		r, err := m.getUser(ctx, &tg.InputUser{UserID: id})
		if err != nil {
			return nil, err
		}
		return r.AsInput(), nil
	}
}
