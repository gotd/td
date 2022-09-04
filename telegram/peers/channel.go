package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

// Channel is channel peer.
type Channel struct {
	raw *tg.Channel
	m   *Manager
}

// Channel creates new Channel, attached to this manager.
func (m *Manager) Channel(u *tg.Channel) Channel {
	m.needsUpdate(channelPeerID(u.ID))
	return Channel{
		raw: u,
		m:   m,
	}
}

// GetChannel gets Channel using given tg.InputChannelClass.
func (m *Manager) GetChannel(ctx context.Context, p tg.InputChannelClass) (Channel, error) {
	ch, err := m.getChannel(ctx, p)
	if err != nil {
		return Channel{}, err
	}
	return m.Channel(ch), nil
}

// Raw returns raw *tg.Channel.
func (c Channel) Raw() *tg.Channel {
	return c.raw
}

// ID returns entity ID.
func (c Channel) ID() int64 {
	return c.raw.GetID()
}

// TDLibPeerID returns TDLibPeerID for this entity.
func (c Channel) TDLibPeerID() constant.TDLibPeerID {
	return channelPeerID(c.raw.GetID())
}

// VisibleName returns visible name of peer.
//
// It returns FirstName + " " + LastName for users, and title for chats and channels.
func (c Channel) VisibleName() string {
	return c.raw.GetTitle()
}

// Username returns peer username, if any.
func (c Channel) Username() (string, bool) {
	return c.raw.GetUsername()
}

// Restricted whether this user/chat/channel is restricted.
func (c Channel) Restricted() ([]tg.RestrictionReason, bool) {
	reason, ok := c.raw.GetRestrictionReason()
	return reason, ok || c.raw.GetRestricted()
}

// Verified whether this user/chat/channel is verified by Telegram.
func (c Channel) Verified() bool {
	return c.raw.Verified
}

// Scam whether this user/chat/channel is probably a scam.
func (c Channel) Scam() bool {
	return c.raw.Scam
}

// Fake whether this user/chat/channel was reported by many users as a fake or scam: be
// careful when interacting with it.
func (c Channel) Fake() bool {
	return c.raw.Fake
}

// InputPeer returns input peer for this peer.
func (c Channel) InputPeer() tg.InputPeerClass {
	return &tg.InputPeerChannel{
		ChannelID:  c.raw.ID,
		AccessHash: c.raw.AccessHash,
	}
}

// Sync updates current object.
func (c Channel) Sync(ctx context.Context) error {
	raw, err := c.m.updateChannel(ctx, c.raw.AsInput())
	if err != nil {
		return errors.Wrap(err, "get channel")
	}
	*c.raw = *raw
	return nil
}

// Manager returns attached Manager.
func (c Channel) Manager() *Manager {
	return c.m
}

// Report reports a peer for violation of telegram's Terms of Service.
func (c Channel) Report(ctx context.Context, reason tg.ReportReasonClass, message string) error {
	if _, err := c.m.api.AccountReportPeer(ctx, &tg.AccountReportPeerRequest{
		Peer:    c.InputPeer(),
		Reason:  reason,
		Message: message,
	}); err != nil {
		return errors.Wrap(err, "report")
	}
	return nil
}

// Photo returns peer photo, if any.
func (c Channel) Photo(ctx context.Context) (*tg.Photo, bool, error) {
	full, err := c.FullRaw(ctx)
	if err != nil {
		return nil, false, err
	}

	p, ok := full.ChatPhoto.AsNotEmpty()
	return p, ok, nil
}

// FullRaw returns *tg.ChannelFull for this Channel.
func (c Channel) FullRaw(ctx context.Context) (*tg.ChannelFull, error) {
	return c.m.getChannelFull(ctx, c.InputChannel())
}

// ToBroadcast tries to convert this Channel to Broadcast.
func (c Channel) ToBroadcast() (Broadcast, bool) {
	if !c.IsBroadcast() {
		return Broadcast{}, false
	}
	return Broadcast{
		Channel: c,
	}, true
}

// IsBroadcast whether this Channel is Broadcast.
func (c Channel) IsBroadcast() bool {
	return c.raw.Broadcast
}

// ToSupergroup tries to convert this Channel to Supergroup.
func (c Channel) ToSupergroup() (Supergroup, bool) {
	if !c.IsSupergroup() {
		return Supergroup{}, false
	}
	return Supergroup{
		Channel: c,
	}, true
}

// IsSupergroup whether this Channel is Supergroup.
func (c Channel) IsSupergroup() bool {
	return c.raw.Megagroup
}

// InviteLinks returns InviteLinks for this peer.
func (c Channel) InviteLinks() InviteLinks {
	return InviteLinks{
		peer: c,
		m:    c.m,
	}
}

// InputChannel returns input user for this user.
func (c Channel) InputChannel() tg.InputChannelClass {
	return &tg.InputChannel{
		ChannelID:  c.raw.ID,
		AccessHash: c.raw.AccessHash,
	}
}

// Creator whether the current user is the creator of this channel.
func (c Channel) Creator() bool {
	return c.raw.Creator
}

// Left whether the current user has left this channel.
func (c Channel) Left() bool {
	return c.raw.Left
}

// HasLink whether this channel has a private join link.
func (c Channel) HasLink() bool {
	return c.raw.HasLink
}

// HasGeo whether this channel has a geoposition.
func (c Channel) HasGeo() bool {
	return c.raw.HasGeo
}

// CallActive whether a group call or livestream is currently active.
func (c Channel) CallActive() bool {
	return c.raw.CallActive
}

// CallNotEmpty whether there's anyone in the group call or livestream.
func (c Channel) CallNotEmpty() bool {
	return c.raw.CallNotEmpty
}

// NoForwards whether that message forwarding from this channel is not allowed.
func (c Channel) NoForwards() bool {
	return c.raw.Noforwards
}

// AdminRights returns admin rights of the user in this channel.
//
// See https://core.telegram.org/api/rights.
func (c Channel) AdminRights() (tg.ChatAdminRights, bool) {
	// TODO(tdakkota): add wrapper for raw object?
	return c.raw.GetAdminRights()
}

// BannedRights returns banned rights of the user in this channel.
//
// See https://core.telegram.org/api/rights.
func (c Channel) BannedRights() (tg.ChatBannedRights, bool) {
	// TODO(tdakkota): add wrapper for raw object?
	return c.raw.GetBannedRights()
}

// DefaultBannedRights returns default chat rights.
//
// See https://core.telegram.org/api/rights.
func (c Channel) DefaultBannedRights() (tg.ChatBannedRights, bool) {
	// TODO(tdakkota): add wrapper for raw object?
	return c.raw.GetDefaultBannedRights()
}

// ParticipantsCount returns count of participants.
func (c Channel) ParticipantsCount() int {
	v, _ := c.raw.GetParticipantsCount()
	return v
}

// Join joins this channel.
func (c Channel) Join(ctx context.Context) error {
	if _, err := c.m.api.ChannelsJoinChannel(ctx, c.InputChannel()); err != nil {
		return errors.Wrap(err, "join channel")
	}
	return nil
}

// Delete deletes this channel.
func (c Channel) Delete(ctx context.Context) error {
	if _, err := c.m.api.ChannelsDeleteChannel(ctx, c.InputChannel()); err != nil {
		return errors.Wrap(err, "delete channel")
	}
	return nil
}

// Leave leaves this channel.
func (c Channel) Leave(ctx context.Context) error {
	if _, err := c.m.api.ChannelsLeaveChannel(ctx, c.InputChannel()); err != nil {
		return errors.Wrap(err, "leave channel")
	}
	return nil
}

// SetTitle sets new title for this Chat.
func (c Channel) SetTitle(ctx context.Context, title string) error {
	if _, err := c.m.api.ChannelsEditTitle(ctx, &tg.ChannelsEditTitleRequest{
		Channel: c.InputChannel(),
		Title:   title,
	}); err != nil {
		return errors.Wrap(err, "edit channel title")
	}
	return nil
}

// SetDescription sets new description for this Chat.
func (c Channel) SetDescription(ctx context.Context, about string) error {
	return c.m.editAbout(ctx, c.InputPeer(), about)
}

// SetReactions sets list of available reactions.
//
// Empty list disables reactions at all.
func (c Channel) SetReactions(ctx context.Context, reactions ...tg.ReactionClass) error {
	return c.m.editReactions(ctx, c.InputPeer(), &tg.ChatReactionsSome{
		Reactions: reactions,
	})
}

// DisableReactions disables reactions.
func (c Channel) DisableReactions(ctx context.Context) error {
	return c.m.editReactions(ctx, c.InputPeer(), &tg.ChatReactionsNone{})
}

// TODO(tdakkota): add more getters, helpers and convertors
