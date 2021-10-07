package peer

import (
	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// Entities is simple peer entities storage.
type Entities struct {
	users    map[int64]*tg.User
	chats    map[int64]*tg.Chat
	channels map[int64]*tg.Channel
}

// NewEntities creates new Entities struct.
func NewEntities(
	users map[int64]*tg.User,
	chats map[int64]*tg.Chat,
	channels map[int64]*tg.Channel,
) Entities {
	return Entities{users: users, chats: chats, channels: channels}
}

// EntitySearchResult is abstraction for different RPC responses which
// contains entities.
type EntitySearchResult interface {
	MapChats() tg.ChatClassArray
	MapUsers() tg.UserClassArray
}

// EntitiesFromResult fills Entities struct using given context.
func EntitiesFromResult(r EntitySearchResult) Entities {
	return NewEntities(
		r.MapUsers().UserToMap(),
		r.MapChats().ChatToMap(),
		r.MapChats().ChannelToMap(),
	)
}

// EntitiesFromUpdate fills Entities struct using given context.
func EntitiesFromUpdate(uctx tg.Entities) Entities {
	return NewEntities(
		uctx.Users,
		uctx.Chats,
		uctx.Channels,
	)
}

// Users returns map of users.
// Notice that returned map is not a copy.
func (ent Entities) Users() map[int64]*tg.User {
	return ent.users
}

// Chats returns map of chats.
// Notice that returned map is not a copy.
func (ent Entities) Chats() map[int64]*tg.Chat {
	return ent.chats
}

// Channels returns map of channels.
// Notice that returned map is not a copy.
func (ent Entities) Channels() map[int64]*tg.Channel {
	return ent.channels
}

// FillFromResult adds and updates all entities from given result.
func (ent Entities) FillFromResult(r EntitySearchResult) {
	r.MapUsers().FillUserMap(ent.users)
	r.MapChats().FillChatMap(ent.chats)
	r.MapChats().FillChannelMap(ent.channels)
}

// FillFromUpdate adds and updates all entities from given updates.
func (ent Entities) FillFromUpdate(uctx tg.Entities) {
	ent.Fill(
		uctx.Users,
		uctx.Chats,
		uctx.Channels,
	)
}

// Fill adds and updates all entities from given maps.
func (ent Entities) Fill(
	users map[int64]*tg.User,
	chats map[int64]*tg.Chat,
	channels map[int64]*tg.Channel,
) {
	for k, v := range users {
		ent.users[k] = v
	}

	for k, v := range chats {
		ent.chats[k] = v
	}

	for k, v := range channels {
		ent.channels[k] = v
	}
}

// ExtractPeer finds and creates InputPeerClass using given PeerClass.
func (ent Entities) ExtractPeer(peerID tg.PeerClass) (tg.InputPeerClass, error) {
	var peer tg.InputPeerClass
	switch p := peerID.(type) {
	case *tg.PeerUser: // peerUser#9db1bc6d
		dialog, ok := ent.users[p.UserID]
		if !ok {
			return nil, xerrors.Errorf("user %d not found", p.UserID)
		}

		peer = &tg.InputPeerUser{
			UserID:     dialog.ID,
			AccessHash: dialog.AccessHash,
		}
	case *tg.PeerChat: // peerChat#bad0e5bb
		dialog, ok := ent.chats[p.ChatID]
		if !ok {
			return nil, xerrors.Errorf("chat %d not found", p.ChatID)
		}

		peer = &tg.InputPeerChat{
			ChatID: dialog.ID,
		}
	case *tg.PeerChannel: // peerChannel#bddde532
		dialog, ok := ent.channels[p.ChannelID]
		if !ok {
			return nil, xerrors.Errorf("channel %d not found", p.ChannelID)
		}

		peer = &tg.InputPeerChannel{
			ChannelID:  dialog.ID,
			AccessHash: dialog.AccessHash,
		}
	default:
		return nil, xerrors.Errorf("unexpected peer type %T", peerID)
	}

	return peer, nil
}

// User finds user by given ID.
func (ent Entities) User(id int64) (*tg.User, bool) {
	v, ok := ent.users[id]
	return v, ok
}

// Chat finds chat by given ID.
func (ent Entities) Chat(id int64) (*tg.Chat, bool) {
	v, ok := ent.chats[id]
	return v, ok
}

// Channel finds channel by given ID.
func (ent Entities) Channel(id int64) (*tg.Channel, bool) {
	v, ok := ent.channels[id]
	return v, ok
}
