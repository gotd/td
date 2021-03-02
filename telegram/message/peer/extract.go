package peer

import (
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// Entities is simple peer entities storage.
type Entities struct {
	users    map[int]*tg.User
	chats    map[int]*tg.Chat
	channels map[int]*tg.Channel
}

// NewEntities creates new Entities struct.
func NewEntities(
	users map[int]*tg.User,
	chats map[int]*tg.Chat,
	channels map[int]*tg.Channel,
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
func EntitiesFromUpdate(uctx tg.UpdateContext) Entities {
	return NewEntities(
		uctx.Users,
		uctx.Chats,
		uctx.Channels,
	)
}

// FillFromResult adds and updates all entities from given result.
func (ent Entities) FillFromResult(r EntitySearchResult) {
	r.MapUsers().FillUserMap(ent.users)
	r.MapChats().FillChatMap(ent.chats)
	r.MapChats().FillChannelMap(ent.channels)
}

// FillFromUpdate adds and updates all entities from given updates.
func (ent Entities) FillFromUpdate(uctx tg.UpdateContext) {
	ent.Fill(
		uctx.Users,
		uctx.Chats,
		uctx.Channels,
	)
}

// Fill adds and updates all entities from given maps.
func (ent Entities) Fill(
	users map[int]*tg.User,
	chats map[int]*tg.Chat,
	channels map[int]*tg.Channel,
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
			return nil, xerrors.Errorf("user %d not found in Update", p.UserID)
		}

		peer = &tg.InputPeerUser{
			UserID:     dialog.ID,
			AccessHash: dialog.AccessHash,
		}
	case *tg.PeerChat: // peerChat#bad0e5bb
		dialog, ok := ent.chats[p.ChatID]
		if !ok {
			return nil, xerrors.Errorf("chat %d not found in Update", p.ChatID)
		}

		peer = &tg.InputPeerChat{
			ChatID: dialog.ID,
		}
	case *tg.PeerChannel: // peerChannel#bddde532
		dialog, ok := ent.channels[p.ChannelID]
		if !ok {
			return nil, xerrors.Errorf("channel %d not found in Update", p.ChannelID)
		}

		peer = &tg.InputPeerChannel{
			ChannelID:  dialog.ID,
			AccessHash: dialog.AccessHash,
		}
	}

	return peer, nil
}
