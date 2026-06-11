package message

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// RichMessage sends a rich message.
//
// A rich message carries structured content (headings, lists, tables, media,
// math and more) instead of a flat string. Build one with the
// telegram/message/rich package, e.g. rich.New(...).Input(), rich.HTML(...) or
// rich.Markdown(...).
func (b *Builder) RichMessage(ctx context.Context, msg tg.InputRichMessageClass) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	req := b.sendRequest(p, "", nil)
	req.RichMessage = msg

	upd, err := b.sender.sendMessage(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "send rich message")
	}

	return upd, nil
}

// RichMessage edits the message, replacing its content with the given rich
// message.
func (b *EditMessageBuilder) RichMessage(ctx context.Context, msg tg.InputRichMessageClass) (tg.UpdatesClass, error) {
	p, err := b.builder.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	req := b.editTextRequest(p, "", nil)
	req.RichMessage = msg

	upd, err := b.builder.sender.editMessage(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "edit rich message")
	}

	return upd, nil
}
