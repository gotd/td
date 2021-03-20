package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
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

// StyledText sends styled text message.
func (b *Builder) StyledText(
	ctx context.Context, text StyledTextOption, texts ...StyledTextOption,
) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	tb := entity.Builder{}
	if err := styling.Perform(&tb, text, texts...); err != nil {
		return nil, err
	}
	msg, entities := tb.Complete()

	upd, err := b.sender.sendMessage(ctx, b.sendRequest(p, msg, entities))
	if err != nil {
		return nil, xerrors.Errorf("send styled text: %w", err)
	}

	return upd, nil
}
