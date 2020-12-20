package telegram

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/td/internal/mt"
)

func (c *Client) ackLoop(ctx context.Context) {
	c.wg.Add(1)
	defer c.wg.Done()

	log := c.log.Named("ack")

	const (
		maxBatchSize      = 20
		forcedSendTimeout = time.Second * 15
	)

	var (
		buf = make([]int64, 0, maxBatchSize)

		// TODO(ernado): remove side-effect.
		ticker = time.NewTicker(forcedSendTimeout)
	)

	send := func() {
		defer func() { buf = buf[:0] }()

		if err := c.writeServiceMessage(ctx, &mt.MsgsAck{MsgIds: buf}); err != nil {
			c.log.Error("Failed to ACK", zap.Error(err))
			return
		}

		log.Debug("ACK", zap.Int64s("message_ids", buf))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if len(buf) > 0 {
				send()
			}
		case msgID := <-c.ackSendChan:
			buf = append(buf, msgID)
			if len(buf) == maxBatchSize {
				send()
				ticker.Reset(forcedSendTimeout)
			}
		}
	}
}
