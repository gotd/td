package message

import (
	"context"
	"crypto/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgmock"
)

func expectSendReport(t *testing.T, reason tg.ReportReasonClass, mock *tgmock.Mock, id int, msg string) {
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesReportRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, reason, req.Reason)
		require.NotZero(t, req.ID)
		require.Equal(t, id, req.ID[0])
		require.Equal(t, msg, req.Message)
	}).ThenTrue()
}

func TestRequestBuilder_Report(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	id64, err := crypto.RandInt64(rand.Reader)
	require.NoError(t, err)
	id := int(id64)
	msg := "abc" + strconv.Itoa(id)

	report := sender.Self().Report(id).Message(msg)

	var r bool
	expectSendReport(t, &tg.InputReportReasonSpam{}, mock, id, msg)
	r, err = report.Spam(ctx)
	require.NoError(t, err)
	require.True(t, r)
	expectSendReport(t, &tg.InputReportReasonViolence{}, mock, id, msg)
	r, err = report.Violence(ctx)
	require.NoError(t, err)
	require.True(t, r)
	expectSendReport(t, &tg.InputReportReasonPornography{}, mock, id, msg)
	r, err = report.Pornography(ctx)
	require.NoError(t, err)
	require.True(t, r)
	expectSendReport(t, &tg.InputReportReasonChildAbuse{}, mock, id, msg)
	r, err = report.ChildAbuse(ctx)
	require.NoError(t, err)
	require.True(t, r)
	expectSendReport(t, &tg.InputReportReasonOther{}, mock, id, msg)
	r, err = report.Other(ctx)
	require.NoError(t, err)
	require.True(t, r)
	expectSendReport(t, &tg.InputReportReasonCopyright{}, mock, id, msg)
	r, err = report.Copyright(ctx)
	require.NoError(t, err)
	require.True(t, r)
	expectSendReport(t, &tg.InputReportReasonGeoIrrelevant{}, mock, id, msg)
	r, err = report.GeoIrrelevant(ctx)
	require.NoError(t, err)
	require.True(t, r)
	expectSendReport(t, &tg.InputReportReasonFake{}, mock, id, msg)
	r, err = report.Fake(ctx)
	require.NoError(t, err)
	require.True(t, r)
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
