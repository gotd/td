package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
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
		return nil, xerrors.Errorf("peer: %w", err)
	}

	upd, err := m.sender.sendScheduledMessages(ctx, &tg.MessagesSendScheduledMessagesRequest{
		Peer: p,
		ID:   append([]int{id}, ids...),
	})
	if err != nil {
		return nil, xerrors.Errorf("send scheduled messages: %w", err)
	}

	return upd, nil
}

// Delete deletes scheduled messages.
func (m *ScheduledManager) Delete(ctx context.Context, id int, ids ...int) (tg.UpdatesClass, error) {
	p, err := m.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	upd, err := m.sender.deleteScheduledMessages(ctx, &tg.MessagesDeleteScheduledMessagesRequest{
		Peer: p,
		ID:   append([]int{id}, ids...),
	})
	if err != nil {
		return nil, xerrors.Errorf("delete scheduled messages: %w", err)
	}

	return upd, nil
}

// Get gets scheduled messages.
func (m *ScheduledManager) Get(ctx context.Context, id int, ids ...int) (tg.ModifiedMessagesMessages, error) {
	p, err := m.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	msgs, err := m.sender.getScheduledMessages(ctx, &tg.MessagesGetScheduledMessagesRequest{
		Peer: p,
		ID:   append([]int{id}, ids...),
	})
	if err != nil {
		return nil, xerrors.Errorf("get scheduled messages: %w", err)
	}

	modified, ok := msgs.AsModified()
	if !ok {
		return nil, xerrors.Errorf("unexpected type %T", msgs)
	}

	return modified, nil
}

// HistoryWithHash gets scheduled messages history.
func (m *ScheduledManager) HistoryWithHash(ctx context.Context, hash int64) (tg.ModifiedMessagesMessages, error) {
	p, err := m.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	msgs, err := m.sender.getScheduledHistory(ctx, &tg.MessagesGetScheduledHistoryRequest{
		Peer: p,
		Hash: hash,
	})
	if err != nil {
		return nil, xerrors.Errorf("get scheduled messages history: %w", err)
	}

	modified, ok := msgs.AsModified()
	if !ok {
		return nil, xerrors.Errorf("unexpected type %T", msgs)
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
