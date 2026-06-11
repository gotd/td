package peers

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// AdminLog is a query for the recent actions log (admin log) of a channel/supergroup.
//
// See https://core.telegram.org/method/channels.getAdminLog.
type AdminLog struct {
	channel Channel

	q      string
	filter tg.ChannelAdminLogEventsFilter
	admins []tg.InputUserClass

	hasFilter bool
}

// AdminLog returns recent actions log query for this channel/supergroup.
//
// Only available to administrators.
func (c Channel) AdminLog() *AdminLog {
	return &AdminLog{channel: c}
}

// Search filters events by the given search query, matching against
// message text and usernames.
func (l *AdminLog) Search(q string) *AdminLog {
	l.q = q
	return l
}

// Filter sets the filter of event types to fetch.
//
// If not set, all event types are returned.
func (l *AdminLog) Filter(filter tg.ChannelAdminLogEventsFilter) *AdminLog {
	l.filter = filter
	l.hasFilter = true
	return l
}

// Admins filters events by the given admins.
//
// If not set, events from all admins are returned.
func (l *AdminLog) Admins(admins ...tg.InputUserClass) *AdminLog {
	l.admins = admins
	return l
}

// AdminLogCallback is the callback called for every admin log event.
type AdminLogCallback = func(event tg.ChannelAdminLogEvent) error

func (l *AdminLog) request(ctx context.Context, maxID, minID int64, limit int) (*tg.ChannelsAdminLogResults, error) {
	req := &tg.ChannelsGetAdminLogRequest{
		Channel: l.channel.InputChannel(),
		Q:       l.q,
		Admins:  l.admins,
		MaxID:   maxID,
		MinID:   minID,
		Limit:   limit,
	}
	if l.hasFilter {
		req.SetEventsFilter(l.filter)
	}

	r, err := l.channel.m.api.ChannelsGetAdminLog(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "get admin log")
	}
	if err := l.channel.m.Apply(ctx, r.Users, r.Chats); err != nil {
		return nil, errors.Wrap(err, "apply entities")
	}
	return r, nil
}

// ForEach calls cb for every event of the admin log, from newest to oldest.
func (l *AdminLog) ForEach(ctx context.Context, cb AdminLogCallback) error {
	const limit = 100

	var maxID int64
	for {
		r, err := l.request(ctx, maxID, 0, limit)
		if err != nil {
			return err
		}
		if len(r.Events) < 1 {
			return nil
		}
		for i, event := range r.Events {
			if err := cb(event); err != nil {
				return errors.Wrapf(err, "callback (index: %d)", i)
			}
			maxID = event.ID
		}
	}
}

// Fetch fetches a single batch of admin log events older than maxID
// (pass zero to start from the newest event).
//
// Use the returned events' IDs as the next maxID to paginate.
func (l *AdminLog) Fetch(ctx context.Context, maxID int64, limit int) ([]tg.ChannelAdminLogEvent, error) {
	r, err := l.request(ctx, maxID, 0, limit)
	if err != nil {
		return nil, err
	}
	return r.Events, nil
}
