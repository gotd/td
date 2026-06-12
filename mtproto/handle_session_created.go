package mtproto

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto"
)

func (c *Conn) handleSessionCreated(b *bin.Buffer) error {
	ctx := context.Background()
	var s mt.NewSessionCreated
	if err := s.Decode(b); err != nil {
		return errors.Wrap(err, "decode")
	}
	c.gotSession.Signal()

	created := proto.MessageID(s.FirstMsgID).Time()
	now := c.clock.Now()
	c.log.Debug(ctx, "Session created",
		log.Int64("unique_id", s.UniqueID),
		log.Int64("first_msg_id", s.FirstMsgID),
		log.Time("first_msg_time", created.Local()),
	)

	if (created.Before(now) && now.Sub(created) > maxPast) || created.Sub(now) > maxFuture {
		c.log.Warn(ctx, "Local clock needs synchronization",
			log.Time("first_msg_time", created),
			log.Time("local", now),
			log.Duration("time_difference", now.Sub(created)),
		)
	}

	c.storeSalt(s.ServerSalt)
	if err := c.handler.OnSession(c.session()); err != nil {
		return errors.Wrap(err, "handler.OnSession")
	}
	return nil
}
