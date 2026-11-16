package telegram

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/gotd/log"

	"github.com/gotd/td/tg"
)

// delayBuckets are histogram buckets in seconds for delays between the
// server and the client. Negative ones are possible without a known server
// time offset, when the local clock is behind the server.
var delayBuckets = metric.WithExplicitBucketBoundaries( // nolint:gochecknoglobals
	-1, -0.1, 0, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 300,
)

// fastBuckets are histogram buckets in seconds for work done inside the
// process, from 100 microseconds.
var fastBuckets = metric.WithExplicitBucketBoundaries( // nolint:gochecknoglobals
	0.0001, 0.00025, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5,
)

// updateTiming is the path of updates pushed by the server:
//
//	message date -> created -> received -> start -> done
//	   (server)      (network)  (client)   (handler)
//
// created is the server clock, the rest is the local one. offset converts
// received to the server clock; it is known from the fastest message
// received (see mtproto.ServerTime), so the receive delay is the time on top
// of the fastest delivery, not the whole network time.
type updateTiming struct {
	created  time.Time     // msg_id time: server created the message with updates
	received time.Time     // encrypted message read from the transport
	offset   time.Duration // server clock minus local clock, zero if unknown
	start    time.Time     // UpdateHandler called
	done     time.Time     // UpdateHandler returned
	failed   bool          // UpdateHandler returned an error
}

// updateTiming records timings of handled updates and logs them at the debug
// level. Neither has peers, message identifiers or content.
func (m *clientMetrics) updateTiming(ctx context.Context, lg log.Helper, t updateTiming, updates tg.UpdatesClass) {
	var (
		receive = t.received.Add(t.offset).Sub(t.created)
		process = t.start.Sub(t.received)
		handle  = t.done.Sub(t.start)
	)
	m.updateReceiveDelay.Record(ctx, receive.Seconds())
	m.updateProcess.Record(ctx, process.Seconds())
	if t.failed {
		m.updateHandle.Record(ctx, handle.Seconds(), m.updateFailed...)
	} else {
		m.updateHandle.Record(ctx, handle.Seconds(), m.updateOK...)
	}

	// Server delays of new messages.
	var (
		messages, channelMessages int
		maxDelay                  time.Duration
	)
	observe := func(channel bool, date int) {
		delay := t.created.Sub(time.Unix(int64(date), 0))
		opts := m.updateOther
		if channel {
			opts = m.updateChannel
			channelMessages++
		}
		m.updateServerDelay.Record(ctx, delay.Seconds(), opts...)
		if messages == 0 || delay > maxDelay {
			maxDelay = delay
		}
		messages++
	}
	observeAll := func(list []tg.UpdateClass) {
		for _, u := range list {
			observeNewMessage(u, observe)
		}
	}
	switch u := updates.(type) {
	case *tg.Updates:
		observeAll(u.Updates)
	case *tg.UpdatesCombined:
		observeAll(u.Updates)
	case *tg.UpdateShort:
		observeNewMessage(u.Update, observe)
	case *tg.UpdateShortMessage:
		observe(false, u.Date)
	case *tg.UpdateShortChatMessage:
		observe(false, u.Date)
	}

	if !lg.Enabled(ctx, log.LevelDebug) {
		return
	}
	attrs := []log.Attr{
		log.Duration("receive_delay", receive),
		log.Duration("process", process),
		log.Duration("handle", handle),
		log.Int("messages", messages),
		log.Int("channel_messages", channelMessages),
	}
	if messages > 0 {
		attrs = append(attrs, log.Duration("server_delay", maxDelay))
	}
	lg.Debug(ctx, "Update timing", attrs...)
}

// observeNewMessage calls fn with the date of the new message of u, if any.
//
// Edits are skipped: their date is the date of the original message.
func observeNewMessage(u tg.UpdateClass, fn func(channel bool, date int)) {
	var (
		msg     tg.MessageClass
		channel bool
	)
	switch u := u.(type) {
	case *tg.UpdateNewChannelMessage:
		msg, channel = u.Message, true
	case *tg.UpdateNewMessage:
		msg = u.Message
	default:
		return
	}
	if m, ok := msg.AsNotEmpty(); ok {
		fn(channel, m.GetDate())
	}
}
