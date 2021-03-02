package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// RequestBuilder is a intermediate builder to make different RPC calls using Sender.
type RequestBuilder struct {
	Builder
}

// ScreenshotNotify sends notification about screenshot.
// Parameter msgID is a ID of message that was screenshotted, can be 0.
func (b *RequestBuilder) ScreenshotNotify(ctx context.Context, msgID int) error {
	p, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	return b.sender.sendScreenshotNotification(ctx, &tg.MessagesSendScreenshotNotificationRequest{
		Peer:         p,
		ReplyToMsgID: msgID,
	})
}

// PeerSettings returns peer settings.
func (b *RequestBuilder) PeerSettings(ctx context.Context) (*tg.PeerSettings, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	return b.sender.getPeerSettings(ctx, p)
}
