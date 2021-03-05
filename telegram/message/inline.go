package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// InlineResult sends inline query result message.
func (b *Builder) InlineResult(ctx context.Context, id string, queryID int64, hideVia bool) error {
	p, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	if err := b.sender.sendInlineBotResult(ctx, &tg.MessagesSendInlineBotResultRequest{
		Silent:       b.silent,
		Background:   b.background,
		ClearDraft:   b.clearDraft,
		HideVia:      hideVia,
		Peer:         p,
		ReplyToMsgID: b.replyToMsgID,
		QueryID:      queryID,
		ID:           id,
		ScheduleDate: b.scheduleDate,
	}); err != nil {
		return xerrors.Errorf("send inline bot result: %w", err)
	}

	return nil
}
