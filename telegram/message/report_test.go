package message

import (
	"context"
	"crypto/rand"
	"strconv"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
)

func expectSendReport(reason tg.ReportReasonClass, mock *rpcmock.Mock, id int, msg string) {
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesReportRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)
		mock.Equal(reason, req.Reason)
		mock.NotEmpty(req.ID)
		mock.Equal(id, req.ID[0])
		mock.Equal(msg, req.Message)
	}).ThenTrue()
}

func TestRequestBuilder_Report(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	id64, err := crypto.RandInt64(rand.Reader)
	mock.NoError(err)
	id := int(id64)
	msg := "abc" + strconv.Itoa(id)

	report := sender.Self().Report(id).Message(msg)

	var r bool
	expectSendReport(&tg.InputReportReasonSpam{}, mock, id, msg)
	r, err = report.Spam(ctx)
	mock.NoError(err)
	mock.True(r)
	expectSendReport(&tg.InputReportReasonViolence{}, mock, id, msg)
	r, err = report.Violence(ctx)
	mock.NoError(err)
	mock.True(r)
	expectSendReport(&tg.InputReportReasonPornography{}, mock, id, msg)
	r, err = report.Pornography(ctx)
	mock.NoError(err)
	mock.True(r)
	expectSendReport(&tg.InputReportReasonChildAbuse{}, mock, id, msg)
	r, err = report.ChildAbuse(ctx)
	mock.NoError(err)
	mock.True(r)
	expectSendReport(&tg.InputReportReasonOther{}, mock, id, msg)
	r, err = report.Other(ctx)
	mock.NoError(err)
	mock.True(r)
	expectSendReport(&tg.InputReportReasonCopyright{}, mock, id, msg)
	r, err = report.Copyright(ctx)
	mock.NoError(err)
	mock.True(r)
	expectSendReport(&tg.InputReportReasonGeoIrrelevant{}, mock, id, msg)
	r, err = report.GeoIrrelevant(ctx)
	mock.NoError(err)
	mock.True(r)
	expectSendReport(&tg.InputReportReasonFake{}, mock, id, msg)
	r, err = report.Fake(ctx)
	mock.NoError(err)
	mock.True(r)
}

func TestRequestBuilder_ReportSpam(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesReportSpamRequest{
		Peer: &tg.InputPeerSelf{},
	}).ThenTrue()

	r, err := sender.Self().ReportSpam(ctx)
	mock.True(r)
	mock.NoError(err)
}
