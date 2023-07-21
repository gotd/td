package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestRequestBuilder_Reaction(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	reaction := []tg.ReactionClass{
		&tg.ReactionEmoji{Emoticon: "A"},
	}
	mock.ExpectCall(&tg.MessagesSendReactionRequest{
		Peer:     &tg.InputPeerSelf{},
		MsgID:    10,
		Reaction: reaction,
	}).ThenResult(&tg.Updates{})
	_, err := sender.Self().Reaction(ctx, 10, reaction[0])
	require.NoError(t, err)

	mock.ExpectCall(&tg.MessagesSendReactionRequest{
		Peer:     &tg.InputPeerSelf{},
		MsgID:    10,
		Reaction: reaction,
	}).ThenRPCErr(testRPCError())
	_, err = sender.Self().Reaction(ctx, 10, reaction[0])
	require.Error(t, err)
}

func TestRequestBuilder_ScreenshotNotify(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendScreenshotNotificationRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, 10, req.ReplyTo.(*tg.InputReplyToMessage).ReplyToMsgID)
	}).ThenResult(&tg.Updates{})
	_, err := sender.Self().ScreenshotNotify(ctx, 10)
	require.NoError(t, err)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendScreenshotNotificationRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, 10, req.ReplyTo.(*tg.InputReplyToMessage).ReplyToMsgID)
	}).ThenRPCErr(testRPCError())
	_, err = sender.Self().ScreenshotNotify(ctx, 10)
	require.Error(t, err)
}

func TestRequestBuilder_PeerSettings(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	expected := tg.PeerSettings{
		ReportSpam: true,
	}
	expected.SetFlags()
	mock.ExpectCall(&tg.MessagesGetPeerSettingsRequest{
		Peer: &tg.InputPeerSelf{},
	}).ThenResult(&tg.MessagesPeerSettings{
		Settings: expected,
	})

	s, err := sender.Self().PeerSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, expected, *s)

	mock.ExpectCall(&tg.MessagesGetPeerSettingsRequest{
		Peer: &tg.InputPeerSelf{},
	}).ThenRPCErr(testRPCError())

	_, err = sender.Self().PeerSettings(ctx)
	require.Error(t, err)
}
