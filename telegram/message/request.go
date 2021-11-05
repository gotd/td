package message

import (
	"context"

	"github.com/ogen-go/errors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// RequestBuilder is an intermediate builder to make different RPC calls using Sender.
type RequestBuilder struct {
	Builder
}

// ScreenshotNotify sends notification about screenshot.
// Parameter msgID is an ID of message that was screenshotted, can be 0.
func (b *RequestBuilder) ScreenshotNotify(ctx context.Context, msgID int) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	upd, err := b.sender.sendScreenshotNotification(ctx, &tg.MessagesSendScreenshotNotificationRequest{
		Peer:         p,
		ReplyToMsgID: msgID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "send screenshot notify")
	}

	return upd, nil
}

// PeerSettings returns peer settings.
func (b *RequestBuilder) PeerSettings(ctx context.Context) (*tg.PeerSettings, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	settings, err := b.sender.getPeerSettings(ctx, p)
	if err != nil {
		return nil, errors.Wrap(err, "get peer settings")
	}

	return settings, nil
}

type startBotBuilder struct {
	bot   tg.InputUserClass
	param string
}

// StartBotOption is an option for StartBot method.
type StartBotOption func(s *startBotBuilder)

// StartBotInputUser sets InputUserClass to start bot.
func StartBotInputUser(bot tg.InputUserClass) func(s *startBotBuilder) {
	return func(s *startBotBuilder) {
		s.bot = bot
	}
}

// StartBotParam sets deeplink param to start bot.
func StartBotParam(param string) func(s *startBotBuilder) {
	return func(s *startBotBuilder) {
		s.param = param
	}
}

// StartBot starts a conversation with a bot using a deep linking parameter.
func (b *RequestBuilder) StartBot(ctx context.Context, opts ...StartBotOption) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	sb := startBotBuilder{}
	for _, opt := range opts {
		opt(&sb)
	}

	if sb.bot == nil {
		user, ok := peer.ToInputUser(p)
		if !ok {
			return nil, errors.Errorf("unexpected peer type %T, try to pass input user manually", p)
		}
		sb.bot = user
	}

	upd, err := b.sender.startBot(ctx, &tg.MessagesStartBotRequest{
		Peer:       p,
		Bot:        sb.bot,
		StartParam: sb.param,
	})
	if err != nil {
		return nil, errors.Wrap(err, "start bot")
	}

	return upd, nil
}
