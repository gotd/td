package tg_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestUpdateDispatcherAny(t *testing.T) {
	ctx := context.Background()
	d := tg.NewUpdateDispatcher()

	var (
		got    []string
		before int
	)
	d.OnBefore(func(context.Context, tg.Entities, tg.UpdateClass) error { before++; return nil })
	d.OnAnyNewMessage(func(_ context.Context, _ tg.Entities, msg tg.MessageClass) error {
		got = append(got, fmt.Sprintf("new %d", msg.GetID()))
		return nil
	})
	d.OnAnyEditMessage(func(_ context.Context, _ tg.Entities, msg tg.MessageClass) error {
		got = append(got, fmt.Sprintf("edit %d", msg.GetID()))
		return nil
	})
	d.OnAnyDeleteMessages(func(_ context.Context, _ tg.Entities, channelID int64, ids []int) error {
		got = append(got, fmt.Sprintf("delete %d %v", channelID, ids))
		return nil
	})
	d.OnAnyReadInbox(func(_ context.Context, _ tg.Entities, peer tg.PeerClass, maxID int) error {
		got = append(got, fmt.Sprintf("inbox %s %d", peer, maxID))
		return nil
	})
	d.OnAnyReadOutbox(func(_ context.Context, _ tg.Entities, peer tg.PeerClass, maxID int) error {
		got = append(got, fmt.Sprintf("outbox %s %d", peer, maxID))
		return nil
	})
	d.OnAnyTyping(func(_ context.Context, _ tg.Entities, peer, from tg.PeerClass, _ tg.SendMessageActionClass) error {
		got = append(got, fmt.Sprintf("typing %s %s", peer, from))
		return nil
	})

	typing := &tg.SendMessageTypingAction{}
	updates := []tg.UpdateClass{
		&tg.UpdateNewMessage{Message: &tg.Message{ID: 1}},
		&tg.UpdateNewChannelMessage{Message: &tg.Message{ID: 2}},
		&tg.UpdateEditMessage{Message: &tg.Message{ID: 3}},
		&tg.UpdateEditChannelMessage{Message: &tg.Message{ID: 4}},
		&tg.UpdateDeleteMessages{Messages: []int{5}},
		&tg.UpdateDeleteChannelMessages{ChannelID: 10, Messages: []int{6}},
		&tg.UpdateReadHistoryInbox{Peer: &tg.PeerUser{UserID: 1}, MaxID: 7},
		&tg.UpdateReadChannelInbox{ChannelID: 10, MaxID: 8},
		&tg.UpdateReadHistoryOutbox{Peer: &tg.PeerChat{ChatID: 2}, MaxID: 9},
		&tg.UpdateReadChannelOutbox{ChannelID: 10, MaxID: 10},
		&tg.UpdateUserTyping{UserID: 1, Action: typing},
		&tg.UpdateChatUserTyping{ChatID: 2, FromID: &tg.PeerUser{UserID: 1}, Action: typing},
		&tg.UpdateChannelUserTyping{ChannelID: 10, FromID: &tg.PeerUser{UserID: 1}, Action: typing},
		// No handler.
		&tg.UpdateConfig{},
	}
	require.NoError(t, d.Handle(ctx, &tg.Updates{Updates: updates}))

	user := (&tg.PeerUser{UserID: 1}).String()
	chat := (&tg.PeerChat{ChatID: 2}).String()
	channel := (&tg.PeerChannel{ChannelID: 10}).String()
	require.Equal(t, []string{
		"new 1",
		"new 2",
		"edit 3",
		"edit 4",
		"delete 0 [5]",
		"delete 10 [6]",
		"inbox " + user + " 7",
		"inbox " + channel + " 8",
		"outbox " + chat + " 9",
		"outbox " + channel + " 10",
		"typing " + user + " " + user,
		"typing " + chat + " " + user,
		"typing " + channel + " " + user,
	}, got)
	require.Equal(t, len(updates), before, "OnBefore is called for every update")

	// Specific handler replaces unified one.
	var specific bool
	d.OnNewChannelMessage(func(context.Context, tg.Entities, *tg.UpdateNewChannelMessage) error {
		specific = true
		return nil
	})
	got = nil
	require.NoError(t, d.Handle(ctx, &tg.Updates{Updates: updates[:2]}))
	require.True(t, specific)
	require.Equal(t, []string{"new 1"}, got)
}

func TestUpdateDispatcherBeforeError(t *testing.T) {
	d := tg.NewUpdateDispatcher()
	stop := errors.New("filtered")
	d.OnBefore(func(_ context.Context, _ tg.Entities, u tg.UpdateClass) error {
		if _, ok := u.(*tg.UpdateNewMessage); ok {
			return stop
		}
		return nil
	})
	var handled []string
	d.OnNewMessage(func(context.Context, tg.Entities, *tg.UpdateNewMessage) error {
		handled = append(handled, "new")
		return nil
	})
	d.OnEditMessage(func(context.Context, tg.Entities, *tg.UpdateEditMessage) error {
		handled = append(handled, "edit")
		return nil
	})

	err := d.Handle(context.Background(), &tg.Updates{Updates: []tg.UpdateClass{
		&tg.UpdateNewMessage{Message: &tg.Message{}},
		&tg.UpdateEditMessage{Message: &tg.Message{}},
	}})
	require.ErrorIs(t, err, stop)
	require.Equal(t, []string{"edit"}, handled, "only filtered update is skipped")
}
