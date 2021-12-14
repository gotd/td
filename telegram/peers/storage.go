package peers

import (
	"context"
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
