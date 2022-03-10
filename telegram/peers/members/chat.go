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
}

func (c *ChatMembers) queryParticipants(ctx context.Context) (*tg.ChatParticipants, error) {
	full, err := c.chat.FullRaw(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get full")
	}
	switch p := full.Participants.(type) {
	case *tg.ChatParticipantsForbidden:
		return nil, &ChatInfoUnavailableError{Info: p}
	case *tg.ChatParticipants:
		return p, nil
	default:
		return nil, errors.Errorf("unexpected type %T", p)
	}
}

// ForEach calls cb for every member of chat.
//
// May return ChatInfoUnavailableError.
func (c *ChatMembers) ForEach(ctx context.Context, cb Callback) error {
	chatDate := time.Unix(int64(c.chat.Raw().Date), 0)
	p, err := c.queryParticipants(ctx)
	if err != nil {
		return errors.Wrap(err, "query")
	}

	for i, participant := range p.Participants {
		userID := participant.GetUserID()
		user, err := c.m.ResolveUserID(ctx, userID)
		if err != nil {
			return errors.Wrapf(err, "get member %d", userID)
		}

		var inviter peers.User
		switch p := participant.(type) {
		case *tg.ChatParticipant:
			inviter, err = c.m.ResolveUserID(ctx, p.InviterID)
		case *tg.ChatParticipantAdmin:
			inviter, err = c.m.ResolveUserID(ctx, p.InviterID)
		}
		if err != nil {
			return errors.Wrap(err, "get inviter")
		}

		if err := cb(ChatMember{
			parent:      c,
			creatorDate: chatDate,
			user:        user,
			inviter:     inviter,
			raw:         participant,
		}); err != nil {
			return errors.Wrapf(err, "callback (index: %d)", i)
		}
	}
	return nil
}

// Count returns total count of members.
func (c *ChatMembers) Count(ctx context.Context) (int, error) {
	p, err := c.queryParticipants(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "query")
	}
	return len(p.Participants), nil
}

// Peer returns chat object.
func (c *ChatMembers) Peer() peers.Peer {
	return c.chat
}

// Kick kicks user member.
//
// If revokeHistory is set, will delete all messages from this member.
func (c *ChatMembers) Kick(ctx context.Context, member tg.InputUserClass, revokeHistory bool) error {
	if _, err := c.m.API().MessagesDeleteChatUser(ctx, &tg.MessagesDeleteChatUserRequest{
		RevokeHistory: revokeHistory,
		ChatID:        c.chat.ID(),
		UserID:        member,
	}); err != nil {
		return errors.Wrapf(err, "delete user (revoke: %v)", revokeHistory)
	}
	return nil
}

// EditRights edits rights of all members in this chat.
func (c *ChatMembers) EditRights(ctx context.Context, options MemberRights) error {
	return editDefaultRights(ctx, c.m.API(), c.chat.InputPeer(), options)
}

// EditAdmin edits admin rights for given user.
func (c *ChatMembers) EditAdmin(ctx context.Context, user tg.InputUserClass, isAdmin bool) error {
	if _, err := c.m.API().MessagesEditChatAdmin(ctx, &tg.MessagesEditChatAdminRequest{
		ChatID:  c.chat.ID(),
		UserID:  user,
		IsAdmin: isAdmin,
	}); err != nil {
		return errors.Wrap(err, "edit admin")
	}
	return nil
}

// Chat returns recent chat members.
func Chat(chat peers.Chat) *ChatMembers {
	return &ChatMembers{
		m:    chat.Manager(),
		chat: chat,
	}
}
