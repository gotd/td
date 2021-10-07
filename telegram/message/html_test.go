package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/telegram/message/entity"
	"github.com/nnqq/td/telegram/message/html"
	"github.com/nnqq/td/tg"
)

func TestHTMLBuilder_String(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	msg := "abc"
	send := "<b>" + msg + "</b>"
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMessageRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, msg, req.Message)
		require.NotZero(t, req.Entities)
		require.Equal(t, &tg.MessageEntityBold{
			Offset: 0,
			Length: entity.ComputeLength(msg),
		}, req.Entities[0])
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().StyledText(ctx, html.Format(nil, "<b>%s</b>", msg))
	require.NoError(t, err)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMessageRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, msg, req.Message)
		require.NotZero(t, req.Entities)
		require.Equal(t, &tg.MessageEntityBold{
			Offset: 0,
			Length: entity.ComputeLength(msg),
		}, req.Entities[0])
	}).ThenRPCErr(testRPCError())

	_, err = sender.Self().StyledText(ctx, html.Bytes(nil, []byte(send)))
	require.Error(t, err)
}
