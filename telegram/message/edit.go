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
func (b *EditMessageBuilder) Text(ctx context.Context, msg string) error {
	p, err := b.builder.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	if err := b.builder.sender.editMessage(ctx,
		b.editTextRequest(p, msg, nil)); err != nil {
		return xerrors.Errorf("edit text message: %w", err)
	}

	return nil
}

// StyledText edits message using given message.
func (b *EditMessageBuilder) StyledText(ctx context.Context, text StyledTextOption, texts ...StyledTextOption) error {
	p, err := b.builder.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	tb := textBuilder{}
	tb.Perform(text, texts...)
	msg, entities := tb.Complete()

	if err := b.builder.sender.editMessage(ctx,
		b.editTextRequest(p, msg, entities)); err != nil {
		return xerrors.Errorf("edit styled text message: %w", err)
	}

	return nil
}

// Media edits message using given media and text.
func (b *EditMessageBuilder) Media(ctx context.Context, media MediaOption) error {
	p, err := b.builder.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	attachment, err := b.builder.applySingleMedia(ctx, p, media)
	if err != nil {
		return err
	}

	req := b.editTextRequest(p, attachment.Message, attachment.Entities)
	req.Media = attachment.Media
	if err := b.builder.sender.editMessage(ctx, req); err != nil {
		return xerrors.Errorf("send media: %w", err)
	}

	return nil
}

// Edit edits message by ID.
func (b *Builder) Edit(id int) *EditMessageBuilder {
	return &EditMessageBuilder{builder: b, id: id}
}
