package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// ReportBuilder is a reporting messages helper.
type ReportBuilder struct {
	sender  *Sender
	peer    peerPromise
	ids     []int
	message string
}

// Message sets additional comment for report.
func (b *ReportBuilder) Message(msg string) *ReportBuilder {
	b.message = msg
	return b
}

func (b *ReportBuilder) send(ctx context.Context, reason tg.ReportReasonClass) (bool, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return false, xerrors.Errorf("peer: %w", err)
	}

	return b.sender.report(ctx, &tg.MessagesReportRequest{
		Peer:    p,
		ID:      b.ids,
		Reason:  reason,
		Message: b.message,
	})
}

// Spam sends report for spam.
func (b *ReportBuilder) Spam(ctx context.Context) (bool, error) {
	return b.send(ctx, &tg.InputReportReasonSpam{})
}

// Violence sends report for violence.
func (b *ReportBuilder) Violence(ctx context.Context) (bool, error) {
	return b.send(ctx, &tg.InputReportReasonViolence{})
}

// Pornography sends report for pornography.
func (b *ReportBuilder) Pornography(ctx context.Context) (bool, error) {
	return b.send(ctx, &tg.InputReportReasonPornography{})
}

// ChildAbuse sends report for child abuse.
func (b *ReportBuilder) ChildAbuse(ctx context.Context) (bool, error) {
	return b.send(ctx, &tg.InputReportReasonChildAbuse{})
}

// Other sends report for other.
func (b *ReportBuilder) Other(ctx context.Context) (bool, error) {
	return b.send(ctx, &tg.InputReportReasonOther{})
}

// Copyright sends report for copyrighted content.
func (b *ReportBuilder) Copyright(ctx context.Context) (bool, error) {
	return b.send(ctx, &tg.InputReportReasonCopyright{})
}

// GeoIrrelevant sends report for irrelevant geogroup.
func (b *ReportBuilder) GeoIrrelevant(ctx context.Context) (bool, error) {
	return b.send(ctx, &tg.InputReportReasonGeoIrrelevant{})
}

// Fake sends report for fake.
func (b *ReportBuilder) Fake(ctx context.Context) (bool, error) {
	return b.send(ctx, &tg.InputReportReasonFake{})
}

// Report reports messages in a chat for violation of Telegram's Terms of Service.
func (b *RequestBuilder) Report(id int, ids ...int) *ReportBuilder {
	return &ReportBuilder{
		sender: b.sender,
		peer:   b.peer,
		ids:    append([]int{id}, ids...),
	}
}

// ReportSpam reports peer for spam.
// NB: You should check that the peer settings of the chat allow us to do that.
func (b *RequestBuilder) ReportSpam(ctx context.Context) (bool, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return false, xerrors.Errorf("peer: %w", err)
	}

	return b.sender.reportSpam(ctx, p)
}
