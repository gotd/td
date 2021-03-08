package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// EditMessageBuilder creates edit message builder.
type EditMessageBuilder struct {
	builder *Builder
	id      int
}

func (b *EditMessageBuilder) editTextRequest(
	p tg.InputPeerClass,
	msg string,
	entities []tg.MessageEntityClass,
) *tg.MessagesEditMessageRequest {
	return &tg.MessagesEditMessageRequest{
		NoWebpage:    b.builder.noWebpage,
		Peer:         p,
		ID:           b.id,
		Message:      msg,
		ReplyMarkup:  b.builder.replyMarkup,
		Entities:     entities,
		ScheduleDate: b.builder.scheduleDate,
	}
}

// Text edits message using given message.
func (b *EditMessageBuilder) Text(ctx context.Context, msg string) (tg.UpdatesClass, error) {
	p, err := b.builder.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	upd, err := b.builder.sender.editMessage(ctx, b.editTextRequest(p, msg, nil))
	if err != nil {
		return nil, xerrors.Errorf("edit styled text message: %w", err)
	}

	return upd, nil
}

// StyledText edits message using given message.
func (b *EditMessageBuilder) StyledText(
	ctx context.Context,
	text StyledTextOption, texts ...StyledTextOption,
) (tg.UpdatesClass, error) {
	p, err := b.builder.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	tb := textBuilder{}
	tb.Perform(text, texts...)
	msg, entities := tb.Complete()

	upd, err := b.builder.sender.editMessage(ctx, b.editTextRequest(p, msg, entities))
	if err != nil {
		return nil, xerrors.Errorf("edit styled text message: %w", err)
	}

	return upd, nil
}

// Media edits message using given media and text.
func (b *EditMessageBuilder) Media(ctx context.Context, media MediaOption) (tg.UpdatesClass, error) {
	p, err := b.builder.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	attachment, err := b.builder.applySingleMedia(ctx, p, media)
	if err != nil {
		return nil, err
	}

	req := b.editTextRequest(p, attachment.Message, attachment.Entities)
	req.Media = attachment.Media

	upd, err := b.builder.sender.editMessage(ctx, req)
	if err != nil {
		return nil, xerrors.Errorf("send media: %w", err)
	}

	return upd, nil
}

// Edit edits message by ID.
func (b *Builder) Edit(id int) *EditMessageBuilder {
	return &EditMessageBuilder{builder: b, id: id}
}
