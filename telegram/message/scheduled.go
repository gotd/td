package message

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// ScheduledManager is a scheduled messages manager.
type ScheduledManager struct {
	peer   peerPromise
	sender *Sender
}

// Send sends scheduled messages.
func (m *ScheduledManager) Send(ctx context.Context, id int, ids ...int) (tg.UpdatesClass, error) {
	p, err := m.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	upd, err := m.sender.sendScheduledMessages(ctx, &tg.MessagesSendScheduledMessagesRequest{
		Peer: p,
		ID:   append([]int{id}, ids...),
	})
	if err != nil {
		return nil, errors.Wrap(err, "send scheduled messages")
	}

	return upd, nil
}

// Delete deletes scheduled messages.
func (m *ScheduledManager) Delete(ctx context.Context, id int, ids ...int) (tg.UpdatesClass, error) {
	p, err := m.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	upd, err := m.sender.deleteScheduledMessages(ctx, &tg.MessagesDeleteScheduledMessagesRequest{
		Peer: p,
		ID:   append([]int{id}, ids...),
	})
	if err != nil {
		return nil, errors.Wrap(err, "delete scheduled messages")
	}

	return upd, nil
}

// Get gets scheduled messages.
func (m *ScheduledManager) Get(ctx context.Context, id int, ids ...int) (tg.ModifiedMessagesMessages, error) {
	p, err := m.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	msgs, err := m.sender.getScheduledMessages(ctx, &tg.MessagesGetScheduledMessagesRequest{
		Peer: p,
		ID:   append([]int{id}, ids...),
	})
	if err != nil {
		return nil, errors.Wrap(err, "get scheduled messages")
	}

	modified, ok := msgs.AsModified()
	if !ok {
		return nil, errors.Errorf("unexpected type %T", msgs)
	}

	return modified, nil
}

// HistoryWithHash gets scheduled messages history.
func (m *ScheduledManager) HistoryWithHash(ctx context.Context, hash int64) (tg.ModifiedMessagesMessages, error) {
	p, err := m.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	msgs, err := m.sender.getScheduledHistory(ctx, &tg.MessagesGetScheduledHistoryRequest{
		Peer: p,
		Hash: hash,
	})
	if err != nil {
		return nil, errors.Wrap(err, "get scheduled messages history")
	}

	modified, ok := msgs.AsModified()
	if !ok {
		return nil, errors.Errorf("unexpected type %T", msgs)
	}

	return modified, nil
}

// History gets scheduled messages history.
func (m *ScheduledManager) History(ctx context.Context) (tg.ModifiedMessagesMessages, error) {
	return m.HistoryWithHash(ctx, 0)
}

// Scheduled creates new ScheduledManager using resolved peer.
func (b *RequestBuilder) Scheduled() *ScheduledManager {
	return &ScheduledManager{
		peer:   b.peer,
		sender: b.sender,
	}
}
