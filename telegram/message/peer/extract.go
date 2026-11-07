package peer

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
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
//
// Min users and channels do not replace full ones.
func (ent Entities) FillFromResult(r EntitySearchResult) {
	ent.Fill(
		r.MapUsers().UserToMap(),
		r.MapChats().ChatToMap(),
		r.MapChats().ChannelToMap(),
	)
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
//
// Min users and channels (see https://core.telegram.org/api/min) do not
// replace full ones, as they have no usable access hash. Instead, the fields
// a min constructor is known to carry are merged into the full one, like
// TDLib does, see MergeMinUser and MergeMinChannel.
func (ent Entities) Fill(
	users map[int64]*tg.User,
	chats map[int64]*tg.Chat,
	channels map[int64]*tg.Channel,
) {
	for k, v := range users {
		if old, ok := ent.users[k]; ok && v.Min && !old.Min {
			v = MergeMinUser(old, v)
		}
		ent.users[k] = v
	}

	for k, v := range chats {
		ent.chats[k] = v
	}

	for k, v := range channels {
		if old, ok := ent.channels[k]; ok && v.Min && !old.Min {
			v = MergeMinChannel(old, v)
		}
		ent.channels[k] = v
	}
}

// MergeMinUser returns a copy of full user updated with fields of min user
// that min constructors carry reliably, following TDLib
// (UserManager::on_get_user).
//
// Access hash, name, usernames, phone and contact state are kept from full.
// Photo is taken from min only if full has ApplyMinPhoto set, and status only
// if full has none. A deleted min copy of self is ignored, as in TDLib.
func MergeMinUser(full, min *tg.User) *tg.User {
	u := *full
	if full.Self && min.Deleted {
		// TDLib ignores self as deleted min user: it is a frozen account
		// seen in a min context, not a deletion.
		return &u
	}

	u.Deleted = min.Deleted
	u.Verified = min.Verified
	u.Premium = min.Premium
	u.Support = min.Support
	u.Scam = min.Scam
	u.Fake = min.Fake
	u.Restricted = min.Restricted
	u.RestrictionReason = min.RestrictionReason

	u.Bot = min.Bot
	u.BotChatHistory = min.BotChatHistory
	u.BotNochats = min.BotNochats
	u.BotInlineGeo = min.BotInlineGeo
	u.BotAttachMenu = min.BotAttachMenu
	u.BotBusiness = min.BotBusiness
	u.BotHasMainApp = min.BotHasMainApp
	u.BotForumView = min.BotForumView
	u.BotForumCanManageTopics = min.BotForumCanManageTopics
	u.BotCanManageBots = min.BotCanManageBots
	u.BotGuestchat = min.BotGuestchat
	u.BotInfoVersion = min.BotInfoVersion
	u.BotInlinePlaceholder = min.BotInlinePlaceholder
	u.BotActiveUsers = min.BotActiveUsers
	u.BotVerificationIcon = min.BotVerificationIcon

	u.ContactRequirePremium = min.ContactRequirePremium
	u.SendPaidMessagesStars = min.SendPaidMessagesStars
	u.EmojiStatus = min.EmojiStatus
	u.Color = min.Color
	u.ProfileColor = min.ProfileColor
	if min.LangCode != "" {
		u.LangCode = min.LangCode
	}
	if min.StoriesMaxID.MaxID > 0 || min.StoriesUnavailable {
		u.StoriesMaxID = min.StoriesMaxID
		u.StoriesUnavailable = min.StoriesUnavailable
	}
	if full.ApplyMinPhoto {
		u.Photo = min.Photo
	}
	if full.Status == nil {
		u.Status = min.Status
	}

	// Flags of the copy describe full, not the merged fields: recompute them,
	// so that getters and encoding match the fields.
	u.Flags, u.Flags2 = 0, 0
	u.SetFlags()

	return &u
}

// MergeMinChannel returns a copy of full channel updated with fields of min
// channel that min constructors carry reliably, following TDLib
// (ChatManager::on_get_channel).
//
// Access hash, participant status (creator, left, admin and banned rights),
// date, signatures, participant count and colors are kept from full, and the
// photo too if the current user is banned from the channel.
func MergeMinChannel(full, min *tg.Channel) *tg.Channel {
	c := *full

	c.Title = min.Title
	c.Username = min.Username
	c.Usernames = min.Usernames
	if !channelBanned(full) {
		// TDLib keeps the photo of a channel the user is banned from.
		c.Photo = min.Photo
	}
	c.EmojiStatus = min.EmojiStatus
	// Colors are not applied from min channels, as in TDLib.

	c.HasGeo = min.HasGeo
	c.HasLink = min.HasLink
	c.Noforwards = min.Noforwards
	c.Forum = min.Forum
	c.ForumTabs = min.ForumTabs
	c.SlowmodeEnabled = min.SlowmodeEnabled
	c.Megagroup = min.Megagroup
	c.Broadcast = min.Broadcast
	c.Gigagroup = min.Gigagroup
	c.Monoforum = min.Monoforum
	c.Verified = min.Verified
	c.Scam = min.Scam
	c.Fake = min.Fake
	c.Autotranslation = min.Autotranslation
	c.BroadcastMessagesAllowed = min.BroadcastMessagesAllowed
	c.JoinToSend = min.JoinToSend
	c.JoinRequest = min.JoinRequest
	c.LinkedMonoforumID = min.LinkedMonoforumID
	c.Level = min.Level
	c.SendPaidMessagesStars = min.SendPaidMessagesStars
	c.Restricted = min.Restricted
	c.RestrictionReason = min.RestrictionReason
	c.BotVerificationIcon = min.BotVerificationIcon
	c.DefaultBannedRights = min.DefaultBannedRights

	// Flags of the copy describe full, not the merged fields: recompute them,
	// so that getters and encoding match the fields.
	c.Flags, c.Flags2 = 0, 0
	c.SetFlags()

	return &c
}

// channelBanned reports whether the current user is banned (kicked) from
// channel.
func channelBanned(c *tg.Channel) bool {
	return c.BannedRights.ViewMessages
}

// ErrMinPeer means that peer is known only as min constructor, which has no
// usable access hash, so input peer can be built only from a message that
// mentions it, see Entities.ExtractPeerFromMessage and
// https://core.telegram.org/api/min.
var ErrMinPeer = errors.New("min peer has no usable access hash")

// ExtractPeer finds and creates InputPeerClass using given PeerClass.
//
// It returns an error wrapping ErrMinPeer if the user or channel is known
// only as a min constructor, as its access hash is not usable for requests.
// Up to v0.163.0 it returned an input peer with the min access hash; use
// ExtractPeerFromMessage to get a usable input peer for min peers.
func (ent Entities) ExtractPeer(peerID tg.PeerClass) (tg.InputPeerClass, error) {
	var peer tg.InputPeerClass
	switch p := peerID.(type) {
	case *tg.PeerUser: // peerUser#9db1bc6d
		dialog, ok := ent.users[p.UserID]
		if !ok {
			return nil, errors.Errorf("user %d not found", p.UserID)
		}

		if dialog.Min {
			return nil, errors.Wrapf(ErrMinPeer, "user %d", p.UserID)
		}

		peer = &tg.InputPeerUser{
			UserID:     dialog.ID,
			AccessHash: dialog.AccessHash,
		}
	case *tg.PeerChat: // peerChat#bad0e5bb
		dialog, ok := ent.chats[p.ChatID]
		if !ok {
			return nil, errors.Errorf("chat %d not found", p.ChatID)
		}

		peer = &tg.InputPeerChat{
			ChatID: dialog.ID,
		}
	case *tg.PeerChannel: // peerChannel#bddde532
		dialog, ok := ent.channels[p.ChannelID]
		if !ok {
			return nil, errors.Errorf("channel %d not found", p.ChannelID)
		}

		if dialog.Min {
			return nil, errors.Wrapf(ErrMinPeer, "channel %d", p.ChannelID)
		}

		peer = &tg.InputPeerChannel{
			ChannelID:  dialog.ID,
			AccessHash: dialog.AccessHash,
		}
	default:
		return nil, errors.Errorf("unexpected peer type %T", peerID)
	}

	return peer, nil
}

// OffsetPeer returns input peer to use as offset_peer of paginated requests
// (messages.searchGlobal, messages.getDialogs), never failing.
//
// Like ExtractPeer, but like TDLib (get_input_peer_force) it is built even
// for min users and channels, as the server uses offset_peer only to order
// results, and for unknown peers, with zero access hash. Do not use it for
// other requests: min and zero access hashes are rejected there.
func (ent Entities) OffsetPeer(peerID tg.PeerClass) tg.InputPeerClass {
	p, err := ent.ExtractPeer(peerID)
	if err == nil {
		return p
	}
	// Unknown peers get zero access hash, as in TDLib.
	switch v := peerID.(type) {
	case *tg.PeerUser:
		var hash int64
		if u, ok := ent.users[v.UserID]; ok {
			hash = u.AccessHash
		}
		return &tg.InputPeerUser{UserID: v.UserID, AccessHash: hash}
	case *tg.PeerChat:
		return &tg.InputPeerChat{ChatID: v.ChatID}
	case *tg.PeerChannel:
		var hash int64
		if c, ok := ent.channels[v.ChannelID]; ok {
			hash = c.AccessHash
		}
		return &tg.InputPeerChannel{ChannelID: v.ChannelID, AccessHash: hash}
	}
	return &tg.InputPeerEmpty{}
}

// ExtractPeerFromMessage is like ExtractPeer, but for min users and channels
// returns inputPeerUserFromMessage or inputPeerChannelFromMessage, built from
// message msgID in chat msgPeer that mentions the peer (e.g. its author).
//
// See https://core.telegram.org/api/min.
func (ent Entities) ExtractPeerFromMessage(
	peerID tg.PeerClass,
	msgPeer tg.InputPeerClass,
	msgID int,
) (tg.InputPeerClass, error) {
	p, err := ent.ExtractPeer(peerID)
	if !errors.Is(err, ErrMinPeer) {
		return p, err
	}

	switch v := peerID.(type) {
	case *tg.PeerUser:
		return &tg.InputPeerUserFromMessage{
			Peer:   msgPeer,
			MsgID:  msgID,
			UserID: v.UserID,
		}, nil
	case *tg.PeerChannel:
		return &tg.InputPeerChannelFromMessage{
			Peer:      msgPeer,
			MsgID:     msgID,
			ChannelID: v.ChannelID,
		}, nil
	default:
		return nil, err
	}
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
