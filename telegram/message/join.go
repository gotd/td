package message

import (
	"context"

	"github.com/gotd/td/telegram/message/internal"
	"github.com/gotd/td/tg"
)

// JoinLink joins to private chat using given link or hash.
// Input examples:
//
//  t.me/joinchat/AAAAAAAAAAAAAAAA
//	https://t.me/joinchat/AAAAAAAAAAAAAAAA
//  tg:join?invite=AAAAAAAAAAAAAAAA
//  tg://join?invite=AAAAAAAAAAAAAAAA
//
func (s *Sender) JoinLink(ctx context.Context, link string) (tg.UpdatesClass, error) {
	hash := link
	if internal.IsDeeplinkLike(link) {
		l, err := internal.ExpectDeeplink(link, internal.Join)
		if err != nil {
			return nil, err
		}

		hash = l.Args.Get("invite")
	}

	return s.JoinHash(ctx, hash)
}

// JoinHash joins to private chat using given hash.
func (s *Sender) JoinHash(ctx context.Context, hash string) (tg.UpdatesClass, error) {
	return s.importChatInvite(ctx, hash)
}
