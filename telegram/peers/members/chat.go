package members

import (
	"context"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

// ChatMembers is chat Members.
type ChatMembers struct {
	m    *peers.Manager
	chat peers.Chat
	p    []ChatMember
}

// ChatMember is chat Member.
type ChatMember struct {
	parent      *ChatMembers
	creatorDate time.Time
	user        peers.User
	inviter     peers.User
	raw         tg.ChatParticipantClass
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

// ForEach calls cb for every member of chat.
func (c *ChatMembers) ForEach(ctx context.Context, cb Callback) error {
	for i, p := range c.p {
		if err := cb(p); err != nil {
			return errors.Wrapf(err, "callback (index: %d)", i)
		}
	}
	return nil
}

// Count returns total count of members.
func (c *ChatMembers) Count(ctx context.Context) (int, error) {
	return len(c.p), nil
}

// Chat returns recent chat members.
//
// May return ChatInfoUnavailableError.
func Chat(ctx context.Context, chat peers.Chat) (*ChatMembers, error) {
	full, err := chat.FullRaw(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get full")
	}
	m := chat.Manager()
	chatDate := time.Unix(int64(chat.Raw().Date), 0)

	switch p := full.Participants.(type) {
	case *tg.ChatParticipantsForbidden:
		return nil, &ChatInfoUnavailableError{Info: p}
	case *tg.ChatParticipants:
		members := make([]ChatMember, len(p.Participants))
		result := &ChatMembers{
			m:    m,
			chat: chat,
			p:    members,
		}

		for i, participant := range p.Participants {
			userID := participant.GetUserID()
			user, err := m.ResolveUserID(ctx, userID)
			if err != nil {
				return nil, errors.Wrapf(err, "get member %d", userID)
			}

			var inviter peers.User
			switch p := participant.(type) {
			case *tg.ChatParticipant:
				inviter, err = m.ResolveUserID(ctx, p.InviterID)
			case *tg.ChatParticipantAdmin:
				inviter, err = m.ResolveUserID(ctx, p.InviterID)
			}
			if err != nil {
				return nil, errors.Wrap(err, "get inviter")
			}

			members[i] = ChatMember{
				parent:      result,
				creatorDate: chatDate,
				user:        user,
				inviter:     inviter,
				raw:         participant,
			}
		}

		return result, nil
	default:
		return nil, errors.Errorf("unexpected type %T", p)
	}
}
