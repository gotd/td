package peers

import (
	"context"

	"github.com/gotd/td/tg"
)

// Value is storage value.
type Value struct {
	AccessHash int64
}

// Key is storage key.
type Key struct {
	Prefix string
	ID     int64
}

// Storage is peer storage.
type Storage interface {
	Save(ctx context.Context, key Key, value Value) error
	Find(ctx context.Context, key Key) (value Value, found bool, _ error)

	SavePhone(ctx context.Context, phone string, key Key) error
	FindPhone(ctx context.Context, phone string) (key Key, value Value, found bool, err error)

	GetContactsHash(ctx context.Context) (int64, error)
	SaveContactsHash(ctx context.Context, hash int64) error
}

const (
	usersPrefix   = "users_"
	chatsPrefix   = "chats_"
	channelPrefix = "channel_"
)

// Cache is peer entities cache.
type Cache interface {
	SaveUsers(ctx context.Context, users ...*tg.User) error
	SaveUserFulls(ctx context.Context, users ...*tg.UserFull) error
	FindUser(ctx context.Context, id int64) (*tg.User, bool, error)
	FindUserFull(ctx context.Context, id int64) (*tg.UserFull, bool, error)

	SaveChats(ctx context.Context, chats ...*tg.Chat) error
	SaveChatFulls(ctx context.Context, chats ...*tg.ChatFull) error
	FindChat(ctx context.Context, id int64) (*tg.Chat, bool, error)
	FindChatFull(ctx context.Context, id int64) (*tg.ChatFull, bool, error)

	SaveChannels(ctx context.Context, channels ...*tg.Channel) error
	SaveChannelFulls(ctx context.Context, channels ...*tg.ChannelFull) error
	FindChannel(ctx context.Context, id int64) (*tg.Channel, bool, error)
	FindChannelFull(ctx context.Context, id int64) (*tg.ChannelFull, bool, error)
}

// NoopCache is no-op implementation of Cache.
type NoopCache struct{}

var _ Cache = NoopCache{}

// SaveUsers implements Cache.
func (n NoopCache) SaveUsers(ctx context.Context, users ...*tg.User) error {
	return nil
}

// SaveUserFulls implements Cache.
func (n NoopCache) SaveUserFulls(ctx context.Context, users ...*tg.UserFull) error {
	return nil
}

// FindUser implements Cache.
func (n NoopCache) FindUser(ctx context.Context, id int64) (*tg.User, bool, error) {
	return nil, false, nil
}

// FindUserFull implements Cache.
func (n NoopCache) FindUserFull(ctx context.Context, id int64) (*tg.UserFull, bool, error) {
	return nil, false, nil
}

// SaveChats implements Cache.
func (n NoopCache) SaveChats(ctx context.Context, chats ...*tg.Chat) error {
	return nil
}

// SaveChatFulls implements Cache.
func (n NoopCache) SaveChatFulls(ctx context.Context, chats ...*tg.ChatFull) error {
	return nil
}

// FindChat implements Cache.
func (n NoopCache) FindChat(ctx context.Context, id int64) (*tg.Chat, bool, error) {
	return nil, false, nil
}

// FindChatFull implements Cache.
func (n NoopCache) FindChatFull(ctx context.Context, id int64) (*tg.ChatFull, bool, error) {
	return nil, false, nil
}

// SaveChannels implements Cache.
func (n NoopCache) SaveChannels(ctx context.Context, channels ...*tg.Channel) error {
	return nil
}

// SaveChannelFulls implements Cache.
func (n NoopCache) SaveChannelFulls(ctx context.Context, channels ...*tg.ChannelFull) error {
	return nil
}

// FindChannel implements Cache.
func (n NoopCache) FindChannel(ctx context.Context, id int64) (*tg.Channel, bool, error) {
	return nil, false, nil
}

// FindChannelFull implements Cache.
func (n NoopCache) FindChannelFull(ctx context.Context, id int64) (*tg.ChannelFull, bool, error) {
	return nil, false, nil
}
