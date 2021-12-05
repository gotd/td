package message

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// TypingActionBuilder is a helper to create and send typing actions.
//
// See https://core.telegram.org/method/messages.setTyping.
//
// See https://core.telegram.org/type/SendMessageAction.
type TypingActionBuilder struct {
	sender   *Sender
	peer     peerPromise
	topMsgID int
}

// ThreadID sets thread ID to send.
func (b *TypingActionBuilder) ThreadID(id int) *TypingActionBuilder {
	b.topMsgID = id
	return b
}

// ThreadMsg sets message's ID as thread ID to send.
func (b *TypingActionBuilder) ThreadMsg(msg tg.MessageClass) *TypingActionBuilder {
	return b.ThreadID(msg.GetID())
}

func (b *TypingActionBuilder) send(ctx context.Context, action tg.SendMessageActionClass) error {
	p, err := b.peer(ctx)
	if err != nil {
		return errors.Wrap(err, "peer")
	}

	if err := b.sender.setTyping(ctx, &tg.MessagesSetTypingRequest{
		Peer:     p,
		TopMsgID: b.topMsgID,
		Action:   action,
	}); err != nil {
		return errors.Wrap(err, "set typing")
	}

	return nil
}

// Custom sends given action.
func (b *TypingActionBuilder) Custom(ctx context.Context, action tg.SendMessageActionClass) error {
	return b.send(ctx, action)
}

//go:generate go run github.com/gotd/td/telegram/message/internal/mktyping -output typing.gen.go

// TypingAction creates TypingActionBuilder.
func (b *RequestBuilder) TypingAction() *TypingActionBuilder {
	return &TypingActionBuilder{
		sender: b.sender,
		peer:   b.peer,
	}
}
