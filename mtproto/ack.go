package mtproto

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/mt"
)

func (c *Conn) ackLoop(ctx context.Context) error {
	log := c.log.Named("ack")

	var buf []int64
	send := func() {
		defer func() { buf = buf[:0] }()

		if err := c.writeServiceMessage(ctx, &mt.MsgsAck{MsgIds: buf}); err != nil {
			c.log.Error("Failed to ACK", zap.Error(err))
			return
		}

		log.Debug("ACK", zap.Int64s("msg_ids", buf))
	}

	ticker := time.NewTicker(c.ackInterval) // TODO: remove side-effect
	for {
		select {
		case <-ctx.Done():
			return xerrors.Errorf("acl: %w", ctx.Err())
		case <-ticker.C:
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
