package message

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
	"github.com/nnqq/td/tgmock"
)

func testSender(t *testing.T) (*Sender, *tgmock.Mock) {
	mock := tgmock.NewRequire(t)
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

func expectSendMedia(t *testing.T, attachment tg.InputMediaClass, mock *tgmock.Mock) {
	expectSendMediaAndText(t, attachment, mock, "")
}

func expectSendMediaAndText(
	t *testing.T, attachment tg.InputMediaClass, mock *tgmock.Mock,
	msg string, entities ...tg.MessageEntityClass,
) {
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMediaRequest)
		require.True(t, ok)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, msg, req.Message)
		require.Equal(t, attachment, req.Media)
		require.NotZero(t, req.RandomID)

		require.Equal(t, len(entities), len(req.Entities))
		if len(entities) > 0 {
			require.Equal(t, entities, req.Entities)
		}
	}).ThenResult(&tg.Updates{})
}
