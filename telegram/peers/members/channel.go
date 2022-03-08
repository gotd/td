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
	channel peers.Channel
}

// ChannelMember is channel Member.
type ChannelMember struct {
	parent      *ChannelMembers
	creatorDate time.Time
	user        peers.User
	inviter     peers.User
	raw         tg.ChannelParticipantClass
}

// Status returns member Status.
func (c ChannelMember) Status() Status {
	switch c.raw.(type) {
	case *tg.ChannelParticipant:
		return Plain
	case *tg.ChannelParticipantSelf:
		return Plain
	case *tg.ChannelParticipantCreator:
		return Creator
	case *tg.ChannelParticipantAdmin:
		return Admin
	case *tg.ChannelParticipantBanned:
		return Banned
	case *tg.ChannelParticipantLeft:
		return Left
	default:
		return -1
	}
}

// Rank returns admin "rank".
func (c ChannelMember) Rank() (string, bool) {
	switch p := c.raw.(type) {
	case *tg.ChannelParticipant:
		return "", false
	case *tg.ChannelParticipantSelf:
		return "", false
	case *tg.ChannelParticipantCreator:
		return p.GetRank()
	case *tg.ChannelParticipantAdmin:
		return p.GetRank()
	case *tg.ChannelParticipantBanned:
		return "", false
	case *tg.ChannelParticipantLeft:
		return "", false
	default:
		return "", false
	}
}

// JoinDate returns member join date, if it is available.
func (c ChannelMember) JoinDate() (time.Time, bool) {
	switch p := c.raw.(type) {
	case *tg.ChannelParticipant:
		return time.Unix(int64(p.Date), 0), true
	case *tg.ChannelParticipantSelf:
		return time.Unix(int64(p.Date), 0), true
	case *tg.ChannelParticipantCreator:
		return c.creatorDate, false
	case *tg.ChannelParticipantAdmin:
		return time.Unix(int64(p.Date), 0), true
	case *tg.ChannelParticipantBanned:
		return time.Unix(int64(p.Date), 0), true
	case *tg.ChannelParticipantLeft:
		return time.Time{}, false
	default:
		return time.Time{}, false
	}
}

// InvitedBy returns user that invited this member.
func (c ChannelMember) InvitedBy() (peers.User, bool) {
	switch p := c.raw.(type) {
	case *tg.ChannelParticipant:
		return peers.User{}, false
	case *tg.ChannelParticipantSelf:
		return c.inviter, true
	case *tg.ChannelParticipantCreator:
		return peers.User{}, false
	case *tg.ChannelParticipantAdmin:
		_, has := p.GetInviterID()
		return c.inviter, has
	case *tg.ChannelParticipantBanned:
		return peers.User{}, false
	case *tg.ChannelParticipantLeft:
		return peers.User{}, false
	default:
		return peers.User{}, false
	}
}

// User returns member User object.
func (c ChannelMember) User() peers.User {
	return c.user
}

func (c *ChannelMembers) query(ctx context.Context, offset, limit int) (*tg.ChannelsChannelParticipants, error) {
	raw := c.m.API()
	p, err := raw.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
		Channel: c.channel.InputChannel(),
		Filter:  &tg.ChannelParticipantsRecent{},
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
		for i, participant := range m.Participants {
			var (
				userID    int64
				inviterID int64
				err       error
			)
			switch p := participant.(type) {
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
			member := ChannelMember{
				parent:      c,
				creatorDate: channelDate,
				user:        user,
				inviter:     peers.User{},
				raw:         participant,
			}
			if inviterID != 0 {
				inviter, err := c.m.ResolveUserID(ctx, inviterID)
				if err != nil {
					return errors.Wrapf(err, "get inviter %d", inviterID)
				}
				member.inviter = inviter
			}

			if err := cb(member); err != nil {
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

// Channel returns recent channel members.
func Channel(ctx context.Context, channel peers.Channel) (*ChannelMembers, error) {
	m := channel.Manager()
	return &ChannelMembers{
		m:       m,
		channel: channel,
	}, nil
}
