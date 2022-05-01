package peers

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

func (m *Manager) applyUsers(ctx context.Context, input ...tg.UserClass) error {
	var (
		users []*tg.User
		ids   = make([]constant.TDLibPeerID, 0, 16)
	)
	if len(ids) < len(input) {
		ids = make([]constant.TDLibPeerID, 0, len(input))
	}

	for _, user := range input {
		user, ok := user.(*tg.User)
		if !ok {
			// Got nil or Empty.
			continue
		}
		if user.Min {
			// TODO(tdakkota): call some hook to get actual user if got min (e.g. force gaps to getDifference)
			continue
		}
		users = append(users, user)

		id := user.GetID()
		k := Key{
			Prefix: usersPrefix,
			ID:     id,
		}
		v := Value{
			AccessHash: user.AccessHash,
		}
		if err := m.storage.Save(ctx, k, v); err != nil {
			// FIXME(tdakkota): just log errors?
			return errors.Wrapf(err, "save user %d", user.ID)
		}
		if user.Phone != "" {
			if err := m.storage.SavePhone(ctx, user.Phone, k); err != nil {
				return errors.Wrapf(err, "save user %d", user.ID)
			}
		}
		ids = append(ids, userPeerID(id))
	}

	if err := m.cache.SaveUsers(ctx, users...); err != nil {
		return errors.Wrap(err, "cache users")
	}
	m.updated(ids...)
	return nil
}

func (m *Manager) applyChats(ctx context.Context, input ...tg.ChatClass) error {
	var (
		chats    []*tg.Chat
		channels []*tg.Channel
		ids      = make([]constant.TDLibPeerID, 0, 16)
	)
	if len(ids) < len(input) {
		ids = make([]constant.TDLibPeerID, 0, len(input))
	}

	for _, ch := range input {
		var (
			k Key
			v Value
		)
		// FIXME(tdakkota): check min constructors
		switch ch := ch.(type) {
		case *tg.Chat:
			k.ID = ch.ID
			k.Prefix = chatsPrefix
			chats = append(chats, ch)

			ids = append(ids, chatPeerID(ch.ID))
		case *tg.ChatForbidden:
			k.ID = ch.ID
			k.Prefix = chatsPrefix
		case *tg.Channel:
			k.ID = ch.ID
			v.AccessHash = ch.AccessHash
			k.Prefix = channelPrefix
			channels = append(channels, ch)

			ids = append(ids, channelPeerID(ch.ID))
		case *tg.ChannelForbidden:
			k.ID = ch.ID
			v.AccessHash = ch.AccessHash
			k.Prefix = channelPrefix
		default:
			// Got nil or Empty
			continue
		}

		if err := m.storage.Save(ctx, k, v); err != nil {
			// FIXME(tdakkota): just log errors?
			return errors.Wrapf(err, "save chat %d", k.ID)
		}
	}

	if err := m.cache.SaveChats(ctx, chats...); err != nil {
		return errors.Wrap(err, "cache chats")
	}
	if err := m.cache.SaveChannels(ctx, channels...); err != nil {
		return errors.Wrap(err, "cache channels")
	}
	m.updated(ids...)
	return nil
}

// Apply adds given entities to manager state.
func (m *Manager) Apply(ctx context.Context, users []tg.UserClass, chats []tg.ChatClass) error {
	return m.applyEntities(ctx, users, chats)
}

func (m *Manager) applyEntities(ctx context.Context, users []tg.UserClass, chats []tg.ChatClass) error {
	return multierr.Append(m.applyUsers(ctx, users...), m.applyChats(ctx, chats...))
}

func (m *Manager) applyFullUser(ctx context.Context, user *tg.UserFull) error {
	if user == nil {
		return nil
	}
	m.updatedFull(userPeerID(user.ID))
	return m.cache.SaveUserFulls(ctx, user)
}

func (m *Manager) applyFullChat(ctx context.Context, chat *tg.ChatFull) error {
	if chat == nil {
		return nil
	}
	m.updatedFull(chatPeerID(chat.ID))
	return m.cache.SaveChatFulls(ctx, chat)
}

func (m *Manager) applyFullChannel(ctx context.Context, ch *tg.ChannelFull) error {
	if ch == nil {
		return nil
	}
	m.updatedFull(channelPeerID(ch.ID))
	return m.cache.SaveChannelFulls(ctx, ch)
}

func (m *Manager) applyMessagesChats(
	ctx context.Context,
	all tg.MessagesChatsClass,
) (chats []Chat, channels []Channel, _ error) {
	raw := all.GetChats()
	if err := m.applyChats(ctx, raw...); err != nil {
		return nil, nil, errors.Wrap(err, "apply chats")
	}

	for _, ch := range raw {
		switch ch := ch.(type) {
		case *tg.Chat:
			chats = append(chats, m.Chat(ch))
		case *tg.Channel:
			channels = append(channels, m.Channel(ch))
		}
	}

	return chats, channels, nil
}

func (m *Manager) updateContacts(ctx context.Context) ([]tg.UserClass, error) {
	ch := m.sg.DoChan("_contacts", func() (interface{}, error) {
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

			myID, ok := m.myID()
			if !ok {
				return c.Users, nil
			}

			if err := m.storage.SaveContactsHash(ctx, contactsHash(myID, c)); err != nil {
				return nil, errors.Wrap(err, "update contacts hash")
			}
			return c.Users, nil
		case *tg.ContactsContactsNotModified:
			return nil, nil
		default:
			return nil, errors.Errorf("unexpected type %T", r)
		}
	})

	select {
	case r := <-ch:
		if err := r.Err; err != nil {
			return nil, err
		}
		users, ok := r.Val.([]tg.UserClass)
		if !ok {
			return nil, nil
		}
		return users, nil
	case <-ctx.Done():
		return nil, ctx.Err()
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
