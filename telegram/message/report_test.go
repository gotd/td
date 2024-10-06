package message

import (
	"context"
	"crypto/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func expectSendReport(t *testing.T, option []byte, mock *tgmock.Mock, id int, msg string) {
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesReportRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, option, req.Option)
		require.NotZero(t, req.ID)
		require.Equal(t, id, req.ID[0])
		require.Equal(t, msg, req.Message)
	}).ThenResult(&tg.ReportResultReported{})
}

func TestRequestBuilder_Report(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	id64, err := crypto.RandInt64(rand.Reader)
	require.NoError(t, err)
	id := int(id64)
	msg := "abc" + strconv.Itoa(id)

	report := sender.Self().Report(id).Message(msg)

	option := []byte{1, 2, 3}
	expectSendReport(t, option, mock, id, msg)
	r, err := report.Option(ctx, option)
	require.NoError(t, err)
	require.NotNil(t, r)
}

func TestRequestBuilder_ReportSpam(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesReportSpamRequest{
		Peer: &tg.InputPeerSelf{},
	}).ThenTrue()

	r, err := sender.Self().ReportSpam(ctx)
	require.True(t, r)
	require.NoError(t, err)
}
