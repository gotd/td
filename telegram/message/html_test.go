package message

import (
	"context"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/tg"
)

func TestHTMLBuilder_String(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	msg := "abc"
	send := "<b>" + msg + "</b>"
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMessageRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)
		mock.Equal(msg, req.Message)
		mock.NotEmpty(req.Entities)
		mock.Equal(&tg.MessageEntityBold{
			Offset: 0,
			Length: entity.ComputeLength(msg),
		}, req.Entities[0])
	}).ThenResult(&tg.Updates{})

	_, err := sender.Self().StyledText(ctx, html.String(send))
	mock.NoError(err)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMessageRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)
		mock.Equal(msg, req.Message)
		mock.NotEmpty(req.Entities)
		mock.Equal(&tg.MessageEntityBold{
			Offset: 0,
			Length: entity.ComputeLength(msg),
		}, req.Entities[0])
	}).ThenRPCErr(testRPCError())

	_, err = sender.Self().StyledText(ctx, html.Bytes([]byte(send)))
	mock.Error(err)
}
