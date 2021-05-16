package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgmock"
)

func testSender(t *testing.T) (*Sender, *tgmock.Mock) {
	mock := tgmock.NewMock(t, require.New(t))
	sender := NewSender(tg.NewClient(mock))
	return sender, mock
}

func testRPCError() *tgerr.Error {
	return &tgerr.Error{
		Code:    1337,
		Message: "TEST_ERROR",
		Type:    "TEST_ERROR",
	}
}

func expectSendMedia(attachment tg.InputMediaClass, mock *tgmock.Mock) {
	expectSendMediaAndText(attachment, mock, "")
}

func expectSendMediaAndText(
	attachment tg.InputMediaClass, mock *tgmock.Mock,
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
