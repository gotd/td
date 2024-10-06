package message

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
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

func (b *ReportBuilder) send(ctx context.Context, option []byte) (tg.ReportResultClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	return b.sender.report(ctx, &tg.MessagesReportRequest{
		Peer:    p,
		ID:      b.ids,
		Option:  option,
		Message: b.message,
	})
}

// Option sends report with provided option field.
func (b *ReportBuilder) Option(ctx context.Context, option []byte) (tg.ReportResultClass, error) {
	return b.send(ctx, option)
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
		return false, errors.Wrap(err, "peer")
	}

	return b.sender.reportSpam(ctx, p)
}
