package message

import (
	"context"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestBuilder_InlineResult(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendInlineBotResultRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)
		mock.Equal(int64(10), req.QueryID)
		mock.Equal("10", req.ID)
		mock.True(req.HideVia)
	}).ThenResult(&tg.Updates{})
	mock.NoError(sender.Self().InlineResult(ctx, "10", 10, true))

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendInlineBotResultRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)
		mock.Equal(int64(10), req.QueryID)
		mock.Equal("10", req.ID)
		mock.False(req.HideVia)
	}).ThenRPCErr(testRPCError())
	mock.Error(sender.Self().InlineResult(ctx, "10", 10, false))
}
