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
	var (
		inbox, outbox []tg.AnyRead
		typings       []tg.AnyTyping
	)
	d.OnAnyReadInbox(func(_ context.Context, _ tg.Entities, read tg.AnyRead) error {
		inbox = append(inbox, read)
		return nil
	})
	d.OnAnyReadOutbox(func(_ context.Context, _ tg.Entities, read tg.AnyRead) error {
		outbox = append(outbox, read)
		return nil
	})
	d.OnAnyTyping(func(_ context.Context, _ tg.Entities, typing tg.AnyTyping) error {
		typings = append(typings, typing)
		return nil
	})

	var (
		user    = &tg.PeerUser{UserID: 1}
		chat    = &tg.PeerChat{ChatID: 2}
		channel = &tg.PeerChannel{ChannelID: 10}
		saved   = &tg.PeerUser{UserID: 3}
		typing  = &tg.SendMessageTypingAction{}

		readHistoryInbox     = &tg.UpdateReadHistoryInbox{Peer: user, MaxID: 7, TopMsgID: 70}
		readChannelInbox     = &tg.UpdateReadChannelInbox{ChannelID: 10, MaxID: 8}
		readDiscussionInbox  = &tg.UpdateReadChannelDiscussionInbox{ChannelID: 10, TopMsgID: 80, ReadMaxID: 81}
		readMonoForumInbox   = &tg.UpdateReadMonoForumInbox{ChannelID: 10, SavedPeerID: saved, ReadMaxID: 82}
		readHistoryOutbox    = &tg.UpdateReadHistoryOutbox{Peer: chat, MaxID: 9}
		readChannelOutbox    = &tg.UpdateReadChannelOutbox{ChannelID: 10, MaxID: 10}
		readDiscussionOutbox = &tg.UpdateReadChannelDiscussionOutbox{ChannelID: 10, TopMsgID: 90, ReadMaxID: 91}
		readMonoForumOutbox  = &tg.UpdateReadMonoForumOutbox{ChannelID: 10, SavedPeerID: saved, ReadMaxID: 92}
		userTyping           = &tg.UpdateUserTyping{UserID: 1, TopMsgID: 11, Action: typing}
		chatTyping           = &tg.UpdateChatUserTyping{ChatID: 2, FromID: user, Action: typing}
		channelTyping        = &tg.UpdateChannelUserTyping{ChannelID: 10, TopMsgID: 12, FromID: user, Action: typing}
	)
	updates := []tg.UpdateClass{
		&tg.UpdateNewMessage{Message: &tg.Message{ID: 1}},
		&tg.UpdateNewChannelMessage{Message: &tg.Message{ID: 2}},
		&tg.UpdateEditMessage{Message: &tg.Message{ID: 3}},
		&tg.UpdateEditChannelMessage{Message: &tg.Message{ID: 4}},
		&tg.UpdateDeleteMessages{Messages: []int{5}},
		&tg.UpdateDeleteChannelMessages{ChannelID: 10, Messages: []int{6}},
		readHistoryInbox,
		readChannelInbox,
		readDiscussionInbox,
		readMonoForumInbox,
		readHistoryOutbox,
		readChannelOutbox,
		readDiscussionOutbox,
		readMonoForumOutbox,
		userTyping,
		chatTyping,
		channelTyping,
		// No handler.
		&tg.UpdateConfig{},
	}
	require.NoError(t, d.Handle(ctx, &tg.Updates{Updates: updates}))

	require.Equal(t, []string{
		"new 1",
		"new 2",
		"edit 3",
		"edit 4",
		"delete 0 [5]",
		"delete 10 [6]",
	}, got)
	require.Equal(t, []tg.AnyRead{
		{Peer: user, MaxID: 7, TopMsgID: 70, Update: readHistoryInbox},
		{Peer: channel, MaxID: 8, Update: readChannelInbox},
		{Peer: channel, MaxID: 81, TopMsgID: 80, Update: readDiscussionInbox},
		{Peer: channel, MaxID: 82, SavedPeer: saved, Update: readMonoForumInbox},
	}, inbox)
	require.Equal(t, []tg.AnyRead{
		{Peer: chat, MaxID: 9, Update: readHistoryOutbox},
		{Peer: channel, MaxID: 10, Update: readChannelOutbox},
		{Peer: channel, MaxID: 91, TopMsgID: 90, Update: readDiscussionOutbox},
		{Peer: channel, MaxID: 92, SavedPeer: saved, Update: readMonoForumOutbox},
	}, outbox)
	require.Equal(t, []tg.AnyTyping{
		{Peer: user, From: user, TopMsgID: 11, Action: typing, Update: userTyping},
		{Peer: chat, From: user, Action: typing, Update: chatTyping},
		{Peer: channel, From: user, TopMsgID: 12, Action: typing, Update: channelTyping},
	}, typings)
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
