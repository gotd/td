package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func assertCommonMessageOptions(t *testing.T, gotNoForwards bool, gotSendAs tg.InputPeerClass, gotReplyTo tg.InputReplyToClass) {
	t.Helper()

	require.True(t, gotNoForwards)
	require.Equal(t, testSendAsPeer(), gotSendAs)
	require.Equal(t, testReplyTo(), gotReplyTo)
}

func assertSendAsAndReplyTo(t *testing.T, gotSendAs tg.InputPeerClass, gotReplyTo tg.InputReplyToClass) {
	t.Helper()

	require.Equal(t, testSendAsPeer(), gotSendAs)
	require.Equal(t, testReplyTo(), gotReplyTo)
}

func testSendAsPeer() tg.InputPeerClass {
	return &tg.InputPeerChannel{
		ChannelID:  11,
		AccessHash: 22,
	}
}

func testReplyTo() tg.InputReplyToClass {
	return &tg.InputReplyToMessage{
		ReplyToMsgID: 10,
	}
}

func TestBuilder_CommonOptionsMedia(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	photo := &tg.InputPhoto{ID: 10}

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMediaRequest)
		require.True(t, ok)
		assertCommonMessageOptions(t, req.Noforwards, req.SendAs, req.ReplyTo)
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().NoForwards().SendAs(testSendAsPeer()).Reply(10).Media(ctx, Photo(photo))
	require.NoError(t, err)
}

func TestBuilder_CommonOptionsText(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMessageRequest)
		require.True(t, ok)
		assertCommonMessageOptions(t, req.Noforwards, req.SendAs, req.ReplyTo)
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().NoForwards().SendAs(testSendAsPeer()).Reply(10).Text(ctx, "abc")
	require.NoError(t, err)
}

func TestBuilder_CommonOptionsAlbum(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	photo := &tg.InputPhoto{ID: 10}

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMultiMediaRequest)
		require.True(t, ok)
		assertCommonMessageOptions(t, req.Noforwards, req.SendAs, req.ReplyTo)
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().NoForwards().SendAs(testSendAsPeer()).Reply(10).Album(ctx, Photo(photo), Photo(photo))
	require.NoError(t, err)
}

func TestBuilder_CommonOptionsForward(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesForwardMessagesRequest)
		require.True(t, ok)
		assertCommonMessageOptions(t, req.Noforwards, req.SendAs, req.ReplyTo)
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().NoForwards().SendAs(testSendAsPeer()).Reply(10).
		ForwardIDs(&tg.InputPeerSelf{}, 20).Send(ctx)
	require.NoError(t, err)
}

func TestBuilder_CommonOptionsInlineResult(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendInlineBotResultRequest)
		require.True(t, ok)
		assertSendAsAndReplyTo(t, req.SendAs, req.ReplyTo)
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().SendAs(testSendAsPeer()).Reply(10).InlineResult(ctx, "10", 10, true)
	require.NoError(t, err)
}
