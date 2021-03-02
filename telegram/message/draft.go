package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

func (b *Builder) saveDraftRequest(
	peer tg.InputPeerClass,
	msg string,
	entities []tg.MessageEntityClass,
) *tg.MessagesSaveDraftRequest {
	return &tg.MessagesSaveDraftRequest{
		NoWebpage:    b.noWebpage,
		Peer:         peer,
		ReplyToMsgID: b.replyToMsgID,
		Message:      msg,
		Entities:     entities,
	}
}

// ClearDraft clears draft.
// Also, you can use Clear() builder option with any other message send method.
//
// See https://core.telegram.org/api/drafts#clearing-drafts.
func (b *Builder) ClearDraft(ctx context.Context, msg string) error {
	p, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	return b.sender.saveDraft(ctx, &tg.MessagesSaveDraftRequest{Peer: p})
}

// SaveDraft saves given message as draft.
//
// See https://core.telegram.org/api/drafts#saving-drafts.
func (b *Builder) SaveDraft(ctx context.Context, msg string) error {
	p, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	return b.sender.saveDraft(ctx, b.saveDraftRequest(p, msg, nil))
}

// SaveStyledDraft saves given styled message as draft.
//
// See https://core.telegram.org/api/drafts#saving-drafts.
func (b *Builder) SaveStyledDraft(ctx context.Context, text StyledTextOption, texts ...StyledTextOption) error {
	p, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	tb := textBuilder{}
	tb.Perform(text, texts...)
	msg, entities := tb.Complete()
	return b.sender.saveDraft(ctx, b.saveDraftRequest(p, msg, entities))
}
