package peers

import (
	"context"

	"github.com/gotd/td/tg"
)

// Value is value storage.
type Value struct {
	AccessHash int64
}

// Storage is peer storage.
type Storage interface {
	Save(ctx context.Context, prefix []byte, id int64, value Value) error
	Find(ctx context.Context, prefix []byte, id int64) (value Value, found bool, _ error)

	SavePhone(ctx context.Context, phone string, id int64, value Value) error
	FindPhone(ctx context.Context, phone string) (id int64, value Value, found bool, err error)

	GetContactsHash(ctx context.Context) (int64, error)
	SaveContactsHash(ctx context.Context, hash int64) error
}

var (
	usersPrefix   = []byte("users_")
	chatsPrefix   = []byte("chats_")
	channelPrefix = []byte("channel_")
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
