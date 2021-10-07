package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// ForwardBuilder is a forward request builder.
type ForwardBuilder struct {
	builder     *Builder
	from        tg.InputPeerClass
	ids         []int
	withMyScore bool
}

// WithMyScore sets flag to include your score in the forwarded game.
func (b *ForwardBuilder) WithMyScore() *ForwardBuilder {
	b.withMyScore = true
	return b
}

// Send sends forwarded messages.
func (b *ForwardBuilder) Send(ctx context.Context) (tg.UpdatesClass, error) {
	p, err := b.builder.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	upd, err := b.builder.sender.forwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
		Silent:       b.builder.silent,
		Background:   b.builder.background,
		WithMyScore:  b.withMyScore,
		FromPeer:     b.from,
		ID:           b.ids,
		ToPeer:       p,
		ScheduleDate: b.builder.scheduleDate,
	})
	if err != nil {
		return nil, xerrors.Errorf("send inline bot result: %w", err)
	}

	return upd, nil
}

// ForwardIDs creates builder to forward messages by ID.
func (b *Builder) ForwardIDs(from tg.InputPeerClass, id int, ids ...int) *ForwardBuilder {
	return &ForwardBuilder{
		builder: b,
		from:    from,
		ids:     append([]int{id}, ids...),
	}
}

// ForwardMessages creates builder to forward messages.
func (b *Builder) ForwardMessages(from tg.InputPeerClass, msg tg.MessageClass, m ...tg.MessageClass) *ForwardBuilder {
	r := make([]int, 1+len(m))
	r[0] = msg.GetID()
	for i := range m {
		r[i+1] = m[i].GetID()
	}

	return &ForwardBuilder{
		builder: b,
		from:    from,
		ids:     r,
	}
}
