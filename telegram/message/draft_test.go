package message

import (
	"context"
	"testing"

	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func TestDraft(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesSaveDraftRequest{
		Peer:    &tg.InputPeerSelf{},
		Message: "text",
	}).ThenTrue()
	mock.ExpectCall(&tg.MessagesSaveDraftRequest{
		Peer:    &tg.InputPeerSelf{},
		Message: "styled text",
		Entities: []tg.MessageEntityClass{
			&tg.MessageEntityBold{
				Length: len("styled text"),
			},
		},
	}).ThenTrue()
	mock.ExpectCall(&tg.MessagesSaveDraftRequest{
		Peer: &tg.InputPeerSelf{},
	}).ThenTrue()

	mock.NoError(sender.Self().SaveDraft(ctx, "text"))
	mock.NoError(sender.Self().SaveStyledDraft(ctx, styling.Bold("styled text")))
	mock.NoError(sender.Self().ClearDraft(ctx))
}
