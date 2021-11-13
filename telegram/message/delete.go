package message

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// DeleteBuilder is an intermediate builder to delete messages.
// Unlike RevokeBuilder will keep messages for other users.
type DeleteBuilder struct {
	sender *Sender
}

// Delete creates new DeleteBuilder.
func (s *Sender) Delete() *DeleteBuilder {
	return &DeleteBuilder{sender: s}
}

// Messages deletes messages by given IDs, but keeps it for other users.
//
// NB: Telegram counts message IDs globally for private chats (but not for channels). This method does not check that
// all given message IDs from one chat.
func (b *DeleteBuilder) Messages(ctx context.Context, ids ...int) (*tg.MessagesAffectedMessages, error) {
	r, err := b.sender.deleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
		ID: ids,
	})
	if err != nil {
		return nil, errors.Wrap(err, "delete messages")
	}

	return r, nil
}

// RevokeBuilder is an intermediate builder to delete messages.
// Unlike DeleteBuilder will not keep messages for other users.
type RevokeBuilder struct {
	builder *RequestBuilder
}

// Revoke creates new RevokeBuilder.
func (b *RequestBuilder) Revoke() *RevokeBuilder {
	return &RevokeBuilder{builder: b}
}

// Messages deletes messages by given IDs.
//
// NB: Telegram counts message IDs globally for private chats (but not for channels). This method does not check that
// all given message IDs from one chat.
func (b *RevokeBuilder) Messages(ctx context.Context, ids ...int) (*tg.MessagesAffectedMessages, error) {
	p, err := b.builder.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	ch, isChannel := peer.ToInputChannel(p)
	if isChannel {
		r, err := b.builder.sender.deleteChannelMessages(ctx, &tg.ChannelsDeleteMessagesRequest{
			Channel: ch,
			ID:      ids,
		})
		if err != nil {
			return nil, errors.Wrap(err, "delete channel messages")
		}

		return r, nil
	}

	r, err := b.builder.sender.deleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
		Revoke: true,
		ID:     ids,
	})
	if err != nil {
		return nil, errors.Wrap(err, "delete messages")
	}

	return r, nil
}
