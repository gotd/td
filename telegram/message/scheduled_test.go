package message

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
)

func TestScheduledManager_Send(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesSendScheduledMessagesRequest{
		Peer: &tg.InputPeerSelf{},
		ID:   []int{10},
	}).ThenResult(&tg.Updates{})
	mock.NoError(sender.Self().Scheduled().Send(ctx, 10))

	mock.ExpectCall(&tg.MessagesSendScheduledMessagesRequest{
		Peer: &tg.InputPeerSelf{},
		ID:   []int{10},
	}).ThenRPCErr(testRPCError())
	mock.Error(sender.Self().Scheduled().Send(ctx, 10))
}

func TestScheduledManager_Delete(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesDeleteScheduledMessagesRequest{
		Peer: &tg.InputPeerSelf{},
		ID:   []int{10},
	}).ThenResult(&tg.Updates{})
	mock.NoError(sender.Self().Scheduled().Delete(ctx, 10))

	mock.ExpectCall(&tg.MessagesDeleteScheduledMessagesRequest{
		Peer: &tg.InputPeerSelf{},
		ID:   []int{10},
	}).ThenRPCErr(testRPCError())
	mock.Error(sender.Self().Scheduled().Delete(ctx, 10))
}

func TestScheduledManager_Get(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	expected := &tg.MessagesMessagesSlice{
		Messages: []tg.MessageClass{
			&tg.Message{
				ID:     10,
				PeerID: &tg.PeerUser{
					UserID: 10,
				},
			},
		},
	}

	mock.ExpectCall(&tg.MessagesGetScheduledMessagesRequest{
		Peer: &tg.InputPeerSelf{},
		ID:   []int{10},
	}).ThenResult(expected)
	msgs, err := sender.Self().Scheduled().Get(ctx, 10)
	mock.Equal(expected, msgs)
	mock.NoError(err)

	mock.ExpectCall(&tg.MessagesGetScheduledMessagesRequest{
		Peer: &tg.InputPeerSelf{},
		ID:   []int{10},
	}).ThenRPCErr(testRPCError())
	_, err = sender.Self().Scheduled().Get(ctx, 10)
	mock.Error(err)
}

func TestScheduledManager_History(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	expected := &tg.MessagesMessagesSlice{
		Messages: []tg.MessageClass{
			&tg.Message{
				ID:     10,
				PeerID: &tg.PeerUser{
					UserID: 10,
				},
			},
		},
	}

	mock.ExpectCall(&tg.MessagesGetScheduledHistoryRequest{
		Peer: &tg.InputPeerSelf{},
		Hash: 0,
	}).ThenResult(expected)
	msgs, err := sender.Self().Scheduled().History(ctx)
	mock.Equal(expected, msgs)
	mock.NoError(err)

	mock.ExpectCall(&tg.MessagesGetScheduledHistoryRequest{
		Peer: &tg.InputPeerSelf{},
		Hash: 1,
	}).ThenResult(expected)
	msgs, err = sender.Self().Scheduled().HistoryWithHash(ctx, 1)
	mock.Equal(expected, msgs)
	mock.NoError(err)

	mock.ExpectCall(&tg.MessagesGetScheduledHistoryRequest{
		Peer: &tg.InputPeerSelf{},
		Hash: 0,
	}).ThenRPCErr(testRPCError())
	_, err = sender.Self().Scheduled().History(ctx)
	mock.Error(err)
}
