package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/message/inline"
	"github.com/nnqq/td/tg"
)

// InlineResult is a user method to send bot inline query result message.
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

// InlineUpdate is an abstraction for
type InlineUpdate interface {
	GetQueryID() int64
}

// Inline creates new inline.ResultBuilder using given update.
func (s *Sender) Inline(upd InlineUpdate) *inline.ResultBuilder {
	return inline.New(s.raw, s.rand, upd.GetQueryID())
}
