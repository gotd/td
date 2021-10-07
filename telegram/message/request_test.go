package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
)

func TestPure(t *testing.T) {
	s := Sender{}
	dialog := s.Self()

	b1 := dialog.Reply(1).Reply(2).Reply(1)
	b2 := dialog.Reply(2)
	b3 := dialog

	require.Equal(t, 1, b1.replyToMsgID)
	require.Equal(t, 2, b2.replyToMsgID)
	require.Equal(t, 0, b3.replyToMsgID)
}

func TestRequestBuilder_ScreenshotNotify(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendScreenshotNotificationRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, 10, req.ReplyToMsgID)
	}).ThenResult(&tg.Updates{})
	_, err := sender.Self().ScreenshotNotify(ctx, 10)
	require.NoError(t, err)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendScreenshotNotificationRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, 10, req.ReplyToMsgID)
	}).ThenRPCErr(testRPCError())
	_, err = sender.Self().ScreenshotNotify(ctx, 10)
	require.Error(t, err)
}

func TestRequestBuilder_PeerSettings(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	expected := &tg.PeerSettings{
		ReportSpam: true,
	}
	mock.ExpectCall(&tg.MessagesGetPeerSettingsRequest{
		Peer: &tg.InputPeerSelf{},
	}).ThenResult(expected)

	s, err := sender.Self().PeerSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, expected, s)

	mock.ExpectCall(&tg.MessagesGetPeerSettingsRequest{
		Peer: &tg.InputPeerSelf{},
	}).ThenRPCErr(testRPCError())

	_, err = sender.Self().PeerSettings(ctx)
	require.Error(t, err)
}

func TestRequestBuilder_StartBot(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	var peer tg.InputPeerClass = &tg.InputPeerUser{
		UserID:     10,
		AccessHash: 10,
	}

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesStartBotRequest)
		require.True(t, ok)
		require.Equal(t, peer, req.Peer)
		require.Equal(t, &tg.InputUser{
			UserID:     10,
			AccessHash: 10,
		}, req.Bot)
		require.Equal(t, "abc", req.StartParam)
		require.NotZero(t, req.RandomID)
	}).ThenResult(&tg.Updates{})
	_, err := sender.To(peer).StartBot(ctx, StartBotParam("abc"))
	require.NoError(t, err)

	inputUser := &tg.InputUserEmpty{}
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesStartBotRequest)
		require.True(t, ok)
		require.Equal(t, peer, req.Peer)
		require.Equal(t, inputUser, req.Bot)
		require.NotZero(t, req.RandomID)
	}).ThenResult(&tg.Updates{})
	_, err = sender.To(peer).StartBot(ctx, StartBotInputUser(inputUser))
	require.NoError(t, err)

	// Should not make RPC calls.
	_, err = sender.To(&tg.InputPeerChannel{}).StartBot(ctx)
	require.Error(t, err)

	peerFromMsg := &tg.InputPeerUserFromMessage{
		Peer:   peer,
		UserID: 10,
		MsgID:  10,
	}
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesStartBotRequest)
		require.True(t, ok)
		require.Equal(t, peerFromMsg, req.Peer)
		require.Equal(t, &tg.InputUserFromMessage{
			Peer:   peer,
			MsgID:  10,
			UserID: 10,
		}, req.Bot)
		require.NotZero(t, req.RandomID)
	}).ThenRPCErr(testRPCError())

	_, err = sender.To(peerFromMsg).StartBot(ctx)
	require.Error(t, err)
}
