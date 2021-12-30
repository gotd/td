package message

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// RequestBuilder is an intermediate builder to make different RPC calls using Sender.
type RequestBuilder struct {
	Builder
}

// Reaction sends reaction for given message.
func (b *RequestBuilder) Reaction(ctx context.Context, msgID int, reaction string) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	upd, err := b.sender.sendReaction(ctx, p, msgID, reaction)
	if err != nil {
		return nil, errors.Wrap(err, "send reaction")
	}

	return upd, nil
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

	r, err := b.sender.getPeerSettings(ctx, p)
	if err != nil {
		return nil, errors.Wrap(err, "get peer settings")
	}

	return &r.Settings, nil
}
