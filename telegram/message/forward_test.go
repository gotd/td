package message

import (
	"context"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestBuilder_ForwardIDs(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesForwardMessagesRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.ToPeer)
		mock.Equal(&tg.InputPeerSelf{}, req.FromPeer)
		mock.Len(req.ID, 1)
		mock.Equal(10, req.ID[0])
		mock.True(req.WithMyScore)
	}).ThenResult(&tg.Updates{})
	mock.NoError(sender.Self().ForwardIDs(&tg.InputPeerSelf{}, 10).WithMyScore().Send(ctx))

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesForwardMessagesRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.ToPeer)
		mock.Equal(&tg.InputPeerSelf{}, req.FromPeer)
		mock.Len(req.ID, 1)
		mock.Equal(10, req.ID[0])
		mock.True(req.WithMyScore)
	}).ThenRPCErr(testRPCError())
	mock.Error(sender.Self().ForwardIDs(&tg.InputPeerSelf{}, 10).WithMyScore().Send(ctx))
}
