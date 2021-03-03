package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
)

func testSender(t *testing.T) (*Sender, *rpcmock.Mock) {
	mock := rpcmock.NewMock(t, require.New(t))
	sender := NewSender(tg.NewClient(mock))
	return sender, mock
}

func expectSendMedia(attachment tg.InputMediaClass, mock *rpcmock.Mock) {
	expectSendMediaAndText(attachment, mock, "")
}

func expectSendMediaAndText(
	attachment tg.InputMediaClass, mock *rpcmock.Mock,
	msg string, entities ...tg.MessageEntityClass,
) {
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMediaRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)
		mock.Equal(msg, req.Message)
		mock.Equal(attachment, req.Media)
		mock.NotZero(req.RandomID)

		mock.Equal(len(entities), len(req.Entities))
		if len(entities) > 0 {
			mock.Equal(entities, req.Entities)
		}
	}).ThenResult(&tg.Updates{})
}
