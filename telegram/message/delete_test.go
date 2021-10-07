package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestRequestBuilder_Delete(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesDeleteMessagesRequest{
		ID: []int{1, 2, 3},
	}).ThenResult(&tg.MessagesAffectedMessages{})
	_, err := sender.Delete().Messages(ctx, 1, 2, 3)
	require.NoError(t, err)

	mock.ExpectCall(&tg.MessagesDeleteMessagesRequest{
		ID: []int{1, 2, 3},
	}).ThenRPCErr(testRPCError())
	_, err = sender.Delete().Messages(ctx, 1, 2, 3)
	require.Error(t, err)
}

func TestRequestBuilder_Revoke(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesDeleteMessagesRequest{
		Revoke: true,
		ID:     []int{1, 2, 3},
	}).ThenResult(&tg.MessagesAffectedMessages{})
	_, err := sender.To(&tg.InputPeerChat{ChatID: 10}).Revoke().Messages(ctx, 1, 2, 3)
	require.NoError(t, err)

	mock.ExpectCall(&tg.MessagesDeleteMessagesRequest{
		Revoke: true,
		ID:     []int{1, 2, 3},
	}).ThenRPCErr(testRPCError())
	_, err = sender.To(&tg.InputPeerChat{ChatID: 10}).Revoke().Messages(ctx, 1, 2, 3)
	require.Error(t, err)

	ch := &tg.InputPeerChannel{ChannelID: 10, AccessHash: 10}
	inputCh := &tg.InputChannel{
		ChannelID:  ch.ChannelID,
		AccessHash: ch.AccessHash,
	}
	mock.ExpectCall(&tg.ChannelsDeleteMessagesRequest{
		Channel: inputCh,
		ID:      []int{1, 2, 3},
	}).ThenResult(&tg.MessagesAffectedMessages{})
	_, err = sender.To(ch).Revoke().Messages(ctx, 1, 2, 3)
	require.NoError(t, err)

	mock.ExpectCall(&tg.ChannelsDeleteMessagesRequest{
		Channel: inputCh,
		ID:      []int{1, 2, 3},
	}).ThenRPCErr(testRPCError())
	_, err = sender.To(ch).Revoke().Messages(ctx, 1, 2, 3)
	require.Error(t, err)
}
