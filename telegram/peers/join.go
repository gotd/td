package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/internal/deeplink"
	"github.com/gotd/td/tg"
)

type updateWithChats interface {
	tg.UpdatesClass
	GetChats() []tg.ChatClass
}

var _ = []updateWithChats{
	(*tg.Updates)(nil),
	(*tg.UpdatesCombined)(nil),
}

// JoinLink joins to private chat using given link or hash.
// Input examples:
//
//	t.me/+AAAAAAAAAAAAAAAA
//	https://t.me/+AAAAAAAAAAAAAAAA
//	t.me/joinchat/AAAAAAAAAAAAAAAA
//	https://t.me/joinchat/AAAAAAAAAAAAAAAA
//	tg:join?invite=AAAAAAAAAAAAAAAA
//	tg://join?invite=AAAAAAAAAAAAAAAA
func (m *Manager) JoinLink(ctx context.Context, link string) (Peer, error) {
	l, err := deeplink.Expect(link, deeplink.Join)
	if err != nil {
		return nil, err
	}
	return m.ImportInvite(ctx, l.Args.Get("invite"))
}

// ImportInvite imports given hash invite.
func (m *Manager) ImportInvite(ctx context.Context, hash string) (Peer, error) {
	inviteInfo, err := m.api.MessagesCheckChatInvite(ctx, hash)
	if err != nil {
		return Chat{}, errors.Wrap(err, "check invite")
	}

	switch inviteInfo := inviteInfo.(type) {
	case *tg.ChatInviteAlready:
		return m.extractChat(inviteInfo.GetChat())
	case *tg.ChatInvite:
		if err := m.applyUsers(ctx, inviteInfo.Participants...); err != nil {
			return nil, errors.Wrap(err, "update users")
		}
	case *tg.ChatInvitePeek:
		if err := m.applyChats(ctx, inviteInfo.GetChat()); err != nil {
			return Chat{}, errors.Wrap(err, "update chats")
		}
	default:
		return nil, errors.Errorf("unexpected type %T", inviteInfo)
	}

	ch, err := m.importInvite(ctx, hash)
	if err != nil {
		return nil, err
	}

	return m.extractChat(ch)
}

func (m *Manager) importInvite(ctx context.Context, hash string) (tg.ChatClass, error) {
	u, err := m.api.MessagesImportChatInvite(ctx, hash)
	if err != nil {
		return nil, errors.Wrap(err, "import invite")
	}

	updates, ok := u.(updateWithChats)
	if !ok {
		return nil, errors.Errorf("bad result %T type", u)
	}

	// Do not apply it, update hook already did it.
	chats := updates.GetChats()
	if len(chats) < 1 {
		return nil, errors.New("empty result")
	}

	return chats[0], nil
}

func (m *Manager) extractChat(p tg.ChatClass) (Peer, error) {
	// TODO: handle forbidden.
	switch p := p.(type) {
	case *tg.Chat:
		return m.Chat(p), nil
	case *tg.Channel:
		return m.Channel(p), nil
	default:
		return nil, errors.Errorf("unexpected type %T", p)
	}
}
