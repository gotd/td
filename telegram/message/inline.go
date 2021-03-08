package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// InlineResult sends inline query result message.
func (b *Builder) InlineResult(ctx context.Context, id string, queryID int64, hideVia bool) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	upd, err := b.sender.sendInlineBotResult(ctx, &tg.MessagesSendInlineBotResultRequest{
		Silent:       b.silent,
		Background:   b.background,
		ClearDraft:   b.clearDraft,
		HideVia:      hideVia,
		Peer:         p,
		ReplyToMsgID: b.replyToMsgID,
		QueryID:      queryID,
		ID:           id,
		ScheduleDate: b.scheduleDate,
	})
	if err != nil {
		return nil, xerrors.Errorf("send inline bot result: %w", err)
	}

	return upd, nil
}
