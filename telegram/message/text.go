package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

func (b *Builder) sendRequest(
	peer tg.InputPeerClass,
	msg string,
	entities []tg.MessageEntityClass,
) *tg.MessagesSendMessageRequest {
	return &tg.MessagesSendMessageRequest{
		NoWebpage:    b.noWebpage,
		Silent:       b.silent,
		Background:   b.background,
		ClearDraft:   b.clearDraft,
		Peer:         peer,
		ReplyToMsgID: b.replyToMsgID,
		Message:      msg,
		ReplyMarkup:  b.replyMarkup,
		Entities:     entities,
		ScheduleDate: b.scheduleDate,
	}
}

// Text sends text message.
func (b *Builder) Text(ctx context.Context, msg string) error {
	p, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	return b.sender.sendMessage(ctx, b.sendRequest(p, msg, nil))
}

// StyledText sends styled text message.
func (b *Builder) StyledText(ctx context.Context, text StyledTextOption, texts ...StyledTextOption) error {
	p, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	tb := textBuilder{}
	tb.Perform(text, texts...)
	msg, entities := tb.Complete()
	return b.sender.sendMessage(ctx, b.sendRequest(p, msg, entities))
}
