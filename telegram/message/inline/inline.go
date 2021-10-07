package inline

import (
	"context"
	"io"
	"time"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// ResultBuilder is inline result builder.
type ResultBuilder struct {
	raw    *tg.Client
	random io.Reader
	// Set this flag if the results are composed of media files
	gallery bool
	// Set this flag if results may be cached on the server side only for the user that sent
	// the query. By default, results may be returned to any user who sends the same query
	private bool
	// Unique identifier for the answered query
	queryID int64
	// The maximum amount of time in seconds that the result of the inline query may be
	// cached on the server. Defaults to 300.
	cacheTime int
	// Pass the offset that a client should send in the next query with the same text to
	// receive more results. Pass an empty string if there are no more results or if you
	// don‘t support pagination. Offset length can’t exceed 64 bytes.
	nextOffset string
	// If passed, clients will display a button with specified text that switches the user to
	// a private chat with the bot and sends the bot a start message with a certain parameter.
	switchPm tg.InlineBotSwitchPM
}

// New creates new ResultBuilder.
func New(raw *tg.Client, random io.Reader, queryID int64) *ResultBuilder {
	return &ResultBuilder{raw: raw, random: random, queryID: queryID}
}

// Gallery sets flag if the results are composed of media files.
func (r *ResultBuilder) Gallery(gallery bool) *ResultBuilder {
	r.gallery = gallery
	return r
}

// Private sets flag if results may be cached on the server side only for the user that sent
// the query. By default, results may be returned to any user who sends the same query.
func (r *ResultBuilder) Private(private bool) *ResultBuilder {
	r.private = private
	return r
}

// CacheTime sets the maximum amount of time that the result of the inline query may be
// cached on the server. Server's default is 300 seconds.
func (r *ResultBuilder) CacheTime(cacheTime time.Duration) *ResultBuilder {
	return r.CacheTimeSeconds(int(cacheTime.Seconds()))
}

// CacheTimeSeconds sets the maximum amount of time in seconds that the result of the inline query may be
// cached on the server. Server's default is 300.
func (r *ResultBuilder) CacheTimeSeconds(cacheTime int) *ResultBuilder {
	r.cacheTime = cacheTime
	return r
}

// NextOffset sets offset that a client should send in the next query with the same text to
// receive more results. Pass an empty string if there are no more results or if you
// don‘t support pagination. Offset length can’t exceed 64 bytes.
func (r *ResultBuilder) NextOffset(nextOffset string) *ResultBuilder {
	r.nextOffset = nextOffset
	return r
}

// SwitchPM sets SwitchPm field.
//
// If passed, clients will display a button with specified text that switches the user to
// a private chat with the bot and sends the bot a start message with a certain parameter.
func (r *ResultBuilder) SwitchPM(text, startParam string) *ResultBuilder {
	r.switchPm = tg.InlineBotSwitchPM{
		Text:       text,
		StartParam: startParam,
	}
	return r
}

// Set sets inline results for given query.
func (r *ResultBuilder) Set(ctx context.Context, opts ...ResultOption) (bool, error) {
	res := resultPageBuilder{
		results: nil,
		random:  r.random,
	}

	for idx, opt := range opts {
		if err := opt.apply(&res); err != nil {
			return false, xerrors.Errorf("apply %d option: %w", idx+1, err)
		}
	}

	ok, err := r.raw.MessagesSetInlineBotResults(ctx, &tg.MessagesSetInlineBotResultsRequest{
		Private:    r.private,
		QueryID:    r.queryID,
		Results:    res.results,
		CacheTime:  r.cacheTime,
		NextOffset: r.nextOffset,
		SwitchPm:   r.switchPm,
		Gallery:    r.gallery,
	})
	if err != nil {
		return false, xerrors.Errorf("set inline results: %w", err)
	}

	return ok, nil
}
