package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/internal/deeplink"
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
	if deeplink.IsDeeplinkLike(link) {
		l, err := deeplink.Expect(link, deeplink.Join)
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

func inputChannel(p tg.InputPeerClass) (tg.InputChannelClass, error) {
	var input tg.InputChannelClass
	switch ch := p.(type) {
	case *tg.InputPeerChannel:
		input = &tg.InputChannel{
			ChannelID:  ch.ChannelID,
			AccessHash: ch.AccessHash,
		}
	case *tg.InputPeerChannelFromMessage:
		input = &tg.InputChannelFromMessage{
			Peer:      ch.Peer,
			MsgID:     ch.MsgID,
			ChannelID: ch.ChannelID,
		}
	default:
		return nil, xerrors.Errorf("unexpected peer type %T", ch)
	}

	return input, nil
}

// Join joins resolved channel.
// NB: if resolved peer is not a channel, error will be returned.
func (b *RequestBuilder) Join(ctx context.Context) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	input, err := inputChannel(p)
	if err != nil {
		return nil, err
	}

	return b.sender.joinChannel(ctx, input)
}

// Leave leaves resolved channel.
// NB: if resolved peer is not a channel, error will be returned.
func (b *RequestBuilder) Leave(ctx context.Context) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	input, err := inputChannel(p)
	if err != nil {
		return nil, err
	}

	return b.sender.leaveChannel(ctx, input)
}
