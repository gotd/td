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
func EntitiesFromUpdate(uctx tg.Entities) Entities {
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
func (ent Entities) FillFromUpdate(uctx tg.Entities) {
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

// ExtractUser finds and creates InputPeerUser using given PeerUser.
func (ent Entities) ExtractUser(p *tg.PeerUser) (*tg.InputPeerUser, error) {
	dialog, ok := ent.users[p.UserID]
	if !ok {
		return nil, xerrors.Errorf("user %d not found", p.UserID)
	}

	return &tg.InputPeerUser{
		UserID:     dialog.ID,
		AccessHash: dialog.AccessHash,
	}, nil
}

// ExtractChat finds and creates InputPeerChat using given PeerChat.
func (ent Entities) ExtractChat(p *tg.PeerChat) (*tg.InputPeerChat, error) {
	dialog, ok := ent.chats[p.ChatID]
	if !ok {
		return nil, xerrors.Errorf("chat %d not found", p.ChatID)
	}

	return &tg.InputPeerChat{
		ChatID: dialog.ID,
	}, nil
}

// ExtractChannel finds and creates InputPeerChannel using given PeerChannel.
func (ent Entities) ExtractChannel(p *tg.PeerChannel) (*tg.InputPeerChannel, error) {
	dialog, ok := ent.channels[p.ChannelID]
	if !ok {
		return nil, xerrors.Errorf("channel %d not found", p.ChannelID)
	}

	return &tg.InputPeerChannel{
		ChannelID:  dialog.ID,
		AccessHash: dialog.AccessHash,
	}, nil
}

// ExtractPeer finds and creates InputPeerClass using given PeerClass.
func (ent Entities) ExtractPeer(peerID tg.PeerClass) (tg.InputPeerClass, error) {
	switch p := peerID.(type) {
	case *tg.PeerUser: // peerUser#9db1bc6d
		return ent.ExtractUser(p)
	case *tg.PeerChat: // peerChat#bad0e5bb
		return ent.ExtractChat(p)
	case *tg.PeerChannel: // peerChannel#bddde532
		return ent.ExtractChannel(p)
	default:
		return nil, xerrors.Errorf("unexpected peer type %T", peerID)
	}
}
