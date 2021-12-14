package peers

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"

	"github.com/gotd/td/tg"
)

func (m *Manager) applyUsers(ctx context.Context, users ...tg.UserClass) error {
	for _, user := range users {
		user, ok := user.AsNotEmpty()
		if !ok {
			continue
		}
		id := user.GetID()
		v := Value{
			AccessHash: user.AccessHash,
		}
		if err := m.storage.Save(ctx, usersPrefix, id, v); err != nil {
			// FIXME(tdakkota): just log errors?
			return errors.Wrapf(err, "save user %d", user.ID)
		}
		if err := m.storage.SavePhone(ctx, user.Phone, id, v); err != nil {
			return errors.Wrapf(err, "save user %d", user.ID)
		}
	}

	return nil
}

func (m *Manager) applyChats(ctx context.Context, chats ...tg.ChatClass) error {
	for _, chat := range chats {
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

		if err := m.storage.Save(ctx, prefix, id, v); err != nil {
			// FIXME(tdakkota): just log errors?
			return errors.Wrapf(err, "save chat %d", id)
		}
	}

	return nil
}

func (m *Manager) applyEntities(ctx context.Context, users []tg.UserClass, chats []tg.ChatClass) error {
	return multierr.Append(m.applyUsers(ctx, users...), m.applyChats(ctx, chats...))
}

func (m *Manager) applyFullUser(ctx context.Context, user *tg.UserFull) error {
	// TODO(tdakkota): save to storage.
	return nil
}

func (m *Manager) applyFullChat(ctx context.Context, chat *tg.ChatFull) error {
	// TODO(tdakkota): save to storage.
	return nil
}

func (m *Manager) applyFullChannel(ctx context.Context, ch *tg.ChannelFull) error {
	// TODO(tdakkota): save to storage.
	return nil
}

func (m *Manager) updateContacts(ctx context.Context) ([]tg.UserClass, error) {
	if err := m.phone.Acquire(ctx, 1); err != nil {
		return nil, errors.Wrap(err, "acquire phone")
	}
	defer m.phone.Release(1)

	hash, err := m.storage.GetContactsHash(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get contacts hash")
	}

	r, err := m.api.ContactsGetContacts(ctx, hash)
	if err != nil {
		return nil, errors.Wrap(err, "get contacts")
	}

	switch c := r.(type) {
	case *tg.ContactsContacts:
		if err := m.applyUsers(ctx, c.Users...); err != nil {
			return nil, errors.Wrap(err, "update users")
		}

		me, ok := m.me.Load()
		if !ok {
			return nil, nil
		}

		if err := m.storage.SaveContactsHash(ctx, contactsHash(me.ID, c)); err != nil {
			return nil, errors.Wrap(err, "update contacts hash")
		}
		return c.Users, nil
	case *tg.ContactsContactsNotModified:
		return nil, nil
	default:
		return nil, errors.Errorf("unexpected type %T", r)
	}
}

type vectorHash struct {
	state uint64
}

// See https://github.com/tdlib/td/blob/aa8a4979df8fc56032f134471a2cb939a7b0839f/td/telegram/misc.cpp#L242.
func (h *vectorHash) apply(n uint64) {
	h.state ^= h.state >> 21
	h.state ^= h.state << 35
	h.state ^= h.state >> 4
	h.state += n
}

// See https://github.com/tdlib/td/blob/aa8a4979df8fc56032f134471a2cb939a7b0839f/td/telegram/ContactsManager.cpp#L5125.
func contactsHash(myID int64, contacts *tg.ContactsContacts) int64 {
	contacts.MapUsers().SortStableByID()

	var lesserIDx = len(contacts.Users) - 1
	for i, user := range contacts.Users {
		if user.GetID() < myID {
			lesserIDx = i
			break
		}
	}

	var h vectorHash
	h.apply(uint64(len(contacts.Users)))
	for i, contact := range contacts.Users {
		h.apply(uint64(contact.GetID()))
		if i == lesserIDx {
			h.apply(uint64(myID))
		}
	}

	return int64(h.state)
}
