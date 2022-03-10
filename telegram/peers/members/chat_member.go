package members

import (
	"context"
	"time"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

// ChatMember is chat Member.
type ChatMember struct {
	parent      *ChatMembers
	creatorDate time.Time
	user        peers.User
	inviter     peers.User
	raw         tg.ChatParticipantClass
}

// Raw returns raw member object.
func (c ChatMember) Raw() tg.ChatParticipantClass {
	return c.raw
}

// Status returns member Status.
func (c ChatMember) Status() Status {
	switch c.raw.(type) {
	case *tg.ChatParticipant:
		return Plain
	case *tg.ChatParticipantCreator:
		return Creator
	case *tg.ChatParticipantAdmin:
		return Admin
	default:
		return -1
	}
}

// JoinDate returns member join date, if it is available.
func (c ChatMember) JoinDate() (time.Time, bool) {
	switch p := c.raw.(type) {
	case *tg.ChatParticipant:
		return time.Unix(int64(p.Date), 0), true
	case *tg.ChatParticipantCreator:
		return c.creatorDate, true
	case *tg.ChatParticipantAdmin:
		return time.Unix(int64(p.Date), 0), true
	default:
		return time.Time{}, false
	}
}

// InvitedBy returns user that invited this member.
func (c ChatMember) InvitedBy() (peers.User, bool) {
	switch c.raw.(type) {
	case *tg.ChatParticipant:
		return c.inviter, true
	case *tg.ChatParticipantCreator:
		return peers.User{}, false
	case *tg.ChatParticipantAdmin:
		return c.inviter, true
	default:
		return peers.User{}, false
	}
}

// User returns member User object.
func (c ChatMember) User() peers.User {
	return c.user
}

// Kick kicks this member.
//
// If revokeHistory is set, will delete all messages from this member.
func (c ChatMember) Kick(ctx context.Context, revokeHistory bool) error {
	return c.parent.Kick(ctx, c.user.InputUser(), revokeHistory)
}
