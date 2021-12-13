package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// Chat is chat peer.
type Chat struct {
	raw *tg.Chat
	m   *Manager
}

// Chat creates new Chat, attached to this manager.
func (m *Manager) Chat(u *tg.Chat) Chat {
	return Chat{
		raw: u,
		m:   m,
	}
}

// GetChat gets Chat using given id.
func (m *Manager) GetChat(ctx context.Context, id int64) (Chat, error) {
	r, err := m.api.MessagesGetChats(ctx, []int64{id})
	if err != nil {
		return Chat{}, errors.Wrap(err, "get chats")
	}
	chats := r.GetChats()

	if len(chats) < 1 {
		return Chat{}, errors.Errorf("got empty result for chat %d", id)
	}

	if err := m.applyChats(ctx, chats...); err != nil {
		return Chat{}, errors.Wrap(err, "update chat")
	}

	ch, ok := chats[0].(*tg.Chat)
	if !ok {
		// TODO(tdakkota): get better error for forbidden.
		return Chat{}, errors.Errorf("got unexpected type %T", chats[0])
	}

	return m.Chat(ch), nil
}

// Raw returns raw *tg.Chat.
func (c Chat) Raw() *tg.Chat {
	return c.raw
}

// VisibleName returns visible name of peer.
//
// It returns FirstName + " " + LastName for users, and title for chats and channels.
func (c Chat) VisibleName() string {
	return c.raw.GetTitle()
}

// Username returns peer username, if any.
func (c Chat) Username() (string, bool) {
	return "", false
}

// InputPeer returns input peer for this peer.
func (c Chat) InputPeer() tg.InputPeerClass {
	return c.raw.AsInputPeer()
}

// Restricted whether this user/chat/channel is restricted.
func (c Chat) Restricted() ([]tg.RestrictionReason, bool) {
	return nil, false
}

// Verified whether this user/chat/channel is verified by Telegram.
func (c Chat) Verified() bool {
	return false
}

// Scam whether this user/chat/channel is probably a scam.
func (c Chat) Scam() bool {
	return false
}

// Fake whether this user/chat/channel was reported by many users as a fake or scam: be
// careful when interacting with it.
func (c Chat) Fake() bool {
	return false
}

// Report reports a peer for violation of telegram's Terms of Service.
func (c Chat) Report(ctx context.Context, reason tg.ReportReasonClass, message string) error {
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
func (c Chat) Photo(ctx context.Context) (*tg.Photo, bool, error) {
	r, err := c.m.api.MessagesGetFullChat(ctx, c.raw.GetID())
	if err != nil {
		return nil, false, errors.Wrap(err, "get full chat")
	}

	if err := c.m.applyEntities(ctx, r.GetUsers(), r.GetChats()); err != nil {
		return nil, false, err
	}

	full, ok := r.FullChat.(*tg.ChatFull)
	if !ok {
		return nil, false, errors.Errorf("unexpected type %T", r.FullChat)
	}
	chatPhoto, ok := full.GetChatPhoto()
	if !ok {
		return nil, false, nil
	}

	p, ok := chatPhoto.AsNotEmpty()
	return p, ok, nil
}

// Leave leaves this chat.
//
// Parameter deleteMyMessages denotes to remove the entire chat history of the this user in this chat.
func (c Chat) Leave(ctx context.Context, deleteMyMessages bool) error {
	_, err := c.m.api.MessagesDeleteChatUser(ctx, &tg.MessagesDeleteChatUserRequest{
		RevokeHistory: deleteMyMessages,
		ChatID:        c.raw.GetID(),
		UserID:        &tg.InputUserSelf{},
	})
	return err
}
