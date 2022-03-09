package members

import (
	"time"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

// ChannelMember is channel Member.
type ChannelMember struct {
	parent      *ChannelMembers
	creatorDate time.Time
	user        peers.User
	inviter     peers.User
	raw         tg.ChannelParticipantClass
}

// Raw returns raw member object.
func (c ChannelMember) Raw() tg.ChannelParticipantClass {
	return c.raw
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
		return c.creatorDate, true
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
