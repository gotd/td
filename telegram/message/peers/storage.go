package peers

import (
	"context"

	"github.com/go-faster/errors"

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

func saveUsers(ctx context.Context, s Storage, u ...tg.UserClass) error {
	for _, user := range u {
		user, ok := user.AsNotEmpty()
		if !ok {
			continue
		}
		id := user.GetID()
		v := Value{
			AccessHash: user.AccessHash,
		}
		if err := s.Save(ctx, usersPrefix, id, v); err != nil {
			// TODO(tdakkota): move this method to Manager and just log errors?
			return errors.Wrapf(err, "save user %d", user.ID)
		}
		if err := s.SavePhone(ctx, user.Phone, id, v); err != nil {
			return errors.Wrapf(err, "save user %d", user.ID)
		}
	}

	return nil
}

func saveChats(ctx context.Context, s Storage, u ...tg.ChatClass) error {
	for _, chat := range u {
		var (
			prefix []byte
			id     int64
			v      Value
		)
		switch chat := chat.(type) {
		case *tg.ChatEmpty:
			continue
		case *tg.Chat:
			id = chat.ID
		case *tg.ChatForbidden:
			id = chat.ID
			prefix = chatsPrefix
		case *tg.Channel:
			id = chat.ID
			v.AccessHash = chat.AccessHash
			prefix = channelPrefix
		case *tg.ChannelForbidden:
			id = chat.ID
			v.AccessHash = chat.AccessHash
			prefix = channelPrefix
		}

		if err := s.Save(ctx, prefix, id, v); err != nil {
			// TODO(tdakkota): move this method to Manager and just log errors?
			return errors.Wrapf(err, "save chat %d", id)
		}
	}

	return nil
}
