package members

import (
	"context"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

// ChannelMembers is channel Members.
type ChannelMembers struct {
	m       *peers.Manager
	filter  tg.ChannelParticipantsFilterClass
	channel peers.Channel
}

func (c *ChannelMembers) query(ctx context.Context, offset, limit int) (*tg.ChannelsChannelParticipants, error) {
	raw := c.m.API()
	p, err := raw.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
		Channel: c.channel.InputChannel(),
		Filter:  c.filter,
		Offset:  offset,
		Limit:   limit,
	})
	if err != nil {
		return nil, errors.Wrap(err, "get members")
	}

	m, ok := p.AsModified()
	if !ok {
		return nil, errors.Errorf("unexpected type %T", p)
	}
	if err := c.m.Apply(ctx, m.Users, m.Chats); err != nil {
		return nil, errors.Wrap(err, "apply entities")
	}
	return m, nil
}

// ForEach calls cb for every member of channel.
//
// May return ChannelInfoUnavailableError.
func (c *ChannelMembers) ForEach(ctx context.Context, cb Callback) error {
	const limit = 100

	full, err := c.channel.FullRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "get full")
	}
	if !full.CanViewParticipants {
		return &ChannelInfoUnavailableError{}
	}
	channelDate := time.Unix(int64(c.channel.Raw().Date), 0)

	offset := 0
	for {
		m, err := c.query(ctx, offset, limit)
		if err != nil {
			return errors.Wrap(err, "query")
		}

		if len(m.Participants) < 1 {
			return nil
		}
		for i, member := range m.Participants {
			var (
				userID    int64
				inviterID int64
				err       error
			)
			switch p := member.(type) {
			case *tg.ChannelParticipant:
				userID = p.UserID
			case *tg.ChannelParticipantSelf:
				userID = p.UserID
				inviterID = p.InviterID
			case *tg.ChannelParticipantCreator:
				userID = p.UserID
			case *tg.ChannelParticipantAdmin:
				userID = p.UserID
				inviterID = p.InviterID
			case *tg.ChannelParticipantBanned:
				userPeer, ok := p.Peer.(*tg.PeerUser)
				if !ok {
					return errors.Errorf("unexpected type %T", p.Peer)
				}
				userID = userPeer.UserID
			case *tg.ChannelParticipantLeft:
				userPeer, ok := p.Peer.(*tg.PeerUser)
				if !ok {
					return errors.Errorf("unexpected type %T", p.Peer)
				}
				userID = userPeer.UserID
			default:
				return errors.Errorf("unexpected type %T", p)
			}

			user, err := c.m.ResolveUserID(ctx, userID)
			if err != nil {
				return errors.Wrapf(err, "get member %d", userID)
			}
			chm := ChannelMember{
				parent:      c,
				creatorDate: channelDate,
				user:        user,
				inviter:     peers.User{},
				raw:         member,
			}
			if inviterID != 0 {
				inviter, err := c.m.ResolveUserID(ctx, inviterID)
				if err != nil {
					return errors.Wrapf(err, "get inviter %d", inviterID)
				}
				chm.inviter = inviter
			}

			if err := cb(chm); err != nil {
				return errors.Wrapf(err, "callback (index: %d)", i)
			}
		}

		offset += limit
	}
}

// Count returns total count of members.
func (c *ChannelMembers) Count(ctx context.Context) (int, error) {
	m, err := c.query(ctx, 0, 1)
	if err != nil {
		return 0, errors.Wrap(err, "query")
	}
	return m.Count, nil
}

// Peer returns chat object.
func (c *ChannelMembers) Peer() peers.Peer {
	return c.channel
}

// Kick kicks user member.
//
// Needed for parity with ChatMembers to define common interface.
//
// If revokeHistory is set, will delete all messages from this member.
func (c *ChannelMembers) Kick(ctx context.Context, member tg.InputUserClass, revokeHistory bool) error {
	p := convertInputUserToInputPeer(member)
	if revokeHistory {
		if _, err := c.m.API().ChannelsDeleteParticipantHistory(ctx, &tg.ChannelsDeleteParticipantHistoryRequest{
			Channel:     c.channel.InputChannel(),
			Participant: p,
		}); err != nil {
			return errors.Wrap(err, "revoke history")
		}
	}
	return c.KickMember(ctx, p)
}

// KickMember kicks member.
//
// Unlike Kick, KickMember can be used to kick chat member that uses send-as-channel mode.
func (c *ChannelMembers) KickMember(ctx context.Context, member tg.InputPeerClass) error {
	return c.EditMemberRights(ctx, member, MemberRights{
		DenyViewMessages: true,
	})
}

// EditMemberRights edits member rights in this channel.
func (c *ChannelMembers) EditMemberRights(
	ctx context.Context,
	member tg.InputPeerClass,
	options MemberRights,
) error {
	return c.editMemberRights(ctx, member, options)
}

func (c *ChannelMembers) editMemberRights(ctx context.Context, p tg.InputPeerClass, options MemberRights) error {
	if _, err := c.m.API().ChannelsEditBanned(ctx, &tg.ChannelsEditBannedRequest{
		Channel:      c.channel.InputChannel(),
		Participant:  p,
		BannedRights: options.IntoChatBannedRights(),
	}); err != nil {
		return errors.Wrap(err, "edit member rights")
	}
	return nil
}

// EditRights edits rights of all members in this channel.
func (c *ChannelMembers) EditRights(ctx context.Context, options MemberRights) error {
	return editDefaultRights(ctx, c.m.API(), c.channel.InputPeer(), options)
}

// EditAdminRights edits admin rights of given user in this channel.
func (c *ChannelMembers) EditAdminRights(
	ctx context.Context,
	admin tg.InputUserClass,
	options AdminRights,
) error {
	if _, err := c.m.API().ChannelsEditAdmin(ctx, &tg.ChannelsEditAdminRequest{
		Channel:     c.channel.InputChannel(),
		UserID:      admin,
		AdminRights: options.IntoChatAdminRights(),
		Rank:        options.Rank,
	}); err != nil {
		return errors.Wrap(err, "edit admin rights")
	}
	return nil
}

// Channel returns recent channel members.
func Channel(channel peers.Channel) *ChannelMembers {
	return ChannelQuery{Channel: channel}.Recent()
}
