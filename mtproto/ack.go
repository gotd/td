package mtproto

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/mt"
)

func (c *Conn) ackLoop(ctx context.Context) error {
	logger := c.log.Named("ack")

	var buf []int64
	send := func() {
		defer func() { buf = buf[:0] }()

		start := c.clock.Now()
		if err := c.writeServiceMessage(ctx, &mt.MsgsAck{MsgIDs: buf}); err != nil {
			// A failing ack write is a strong signal that the outgoing half of
			// the connection is broken while the read loop still works.
			c.log.Error(ctx, "Failed to ACK", log.Error(err), log.Any("msg_ids", buf))
			return
		}

		logger.Debug(ctx, "Ack",
			log.Any("msg_ids", buf),
			log.Duration("elapsed", c.clock.Now().Sub(start)),
		)
	}

	ticker := c.clock.Ticker(c.ackInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "acl")
		case <-ticker.C():
			if len(buf) > 0 {
				send()
			}
		case msgID := <-c.ackSendChan:
			buf = append(buf, msgID)
			if len(buf) >= c.ackBatchSize {
				send()
				ticker.Reset(c.ackInterval)
			}
		}
	}
}
