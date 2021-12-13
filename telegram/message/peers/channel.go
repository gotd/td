package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// Channel is channel peer.
type Channel struct {
	raw *tg.Channel
	m   *Manager
}

// Channel creates new Channel, attached to this manager.
func (m *Manager) Channel(u *tg.Channel) Channel {
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

// InputPeer returns input peer for this peer.
func (c Channel) InputPeer() tg.InputPeerClass {
	return c.raw.AsInputPeer()
}

// Restricted whether this user/chat/channel is restricted.
func (c Channel) Restricted() ([]tg.RestrictionReason, bool) {
	reason, ok := c.raw.GetRestrictionReason()
	return reason, ok || c.raw.GetRestricted()
}

// Verified whether this user/chat/channel is verified by Telegram.
func (c Channel) Verified() bool {
	return c.raw.GetVerified()
}

// Scam whether this user/chat/channel is probably a scam.
func (c Channel) Scam() bool {
	return c.raw.GetScam()
}

// Fake whether this user/chat/channel was reported by many users as a fake or scam: be
// careful when interacting with it.
func (c Channel) Fake() bool {
	return c.raw.GetFake()
}

// Sync updates current object.
func (c Channel) Sync(ctx context.Context) error {
	raw, err := c.m.getChannel(ctx, c.raw.AsInput())
	if err != nil {
		return errors.Wrap(err, "get channel")
	}
	*c.raw = *raw
	return nil
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
	r, err := c.m.api.ChannelsGetFullChannel(ctx, c.InputChannel())
	if err != nil {
		return nil, false, errors.Wrap(err, "get full channel")
	}

	if err := c.m.applyEntities(ctx, r.GetUsers(), r.GetChats()); err != nil {
		return nil, false, err
	}

	full, ok := r.FullChat.(*tg.ChannelFull)
	if !ok {
		return nil, false, errors.Errorf("unexpected type %T", r.FullChat)
	}

	p, ok := full.ChatPhoto.AsNotEmpty()
	return p, ok, nil
}

// ToBroadcast tries to convert this Channel to Broadcast.
func (c Channel) ToBroadcast() (Broadcast, bool) {
	if !c.raw.Broadcast {
		return Broadcast{}, false
	}
	return Broadcast{
		Channel: c,
	}, true
}

// TODO(tdakkota): add more getters, helpers and convertors

// InputChannel returns input user for this user.
func (c Channel) InputChannel() tg.InputChannelClass {
	return c.raw.AsInput()
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
