package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/message/entity"
	"github.com/nnqq/td/telegram/message/styling"
	"github.com/nnqq/td/tg"
)

func (b *Builder) sendRequest(
	p tg.InputPeerClass,
	msg string,
	entities []tg.MessageEntityClass,
) *tg.MessagesSendMessageRequest {
	return &tg.MessagesSendMessageRequest{
		NoWebpage:    b.noWebpage,
		Silent:       b.silent,
		Background:   b.background,
		ClearDraft:   b.clearDraft,
		Peer:         p,
		ReplyToMsgID: b.replyToMsgID,
		Message:      msg,
		ReplyMarkup:  b.replyMarkup,
		Entities:     entities,
		ScheduleDate: b.scheduleDate,
	}
}

// Text sends text message.
func (b *Builder) Text(ctx context.Context, msg string) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	upd, err := b.sender.sendMessage(ctx, b.sendRequest(p, msg, nil))
	if err != nil {
		return nil, xerrors.Errorf("send text: %w", err)
	}

	return upd, nil
}

// Textf formats and sends text message.
func (b *Builder) Textf(ctx context.Context, format string, args ...interface{}) (tg.UpdatesClass, error) {
	return b.Text(ctx, formatMessage(format, args...))
}

// StyledText sends styled text message.
func (b *Builder) StyledText(ctx context.Context, texts ...StyledTextOption) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	tb := entity.Builder{}
	if err := styling.Perform(&tb, texts...); err != nil {
		return nil, err
	}
	msg, entities := tb.Complete()

	upd, err := b.sender.sendMessage(ctx, b.sendRequest(p, msg, entities))
	if err != nil {
		return nil, xerrors.Errorf("send styled text: %w", err)
	}

	return upd, nil
}
