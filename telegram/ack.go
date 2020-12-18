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

	const (
		ackMaxBatchSize      = 20
		ackForcedSendTimeout = time.Second * 15
	)

	var (
		buff  = make([]int64, 0, ackMaxBatchSize)
		timer = time.NewTimer(ackForcedSendTimeout)
	)

	sendAcks := func(ctx context.Context) {
		defer func() { buff = buff[:0] }()

		if err := c.writeServiceMessage(ctx, &mt.MsgsAck{MsgIds: buff}); err != nil {
			c.log.Error("send acks", zap.Error(err))
			return
		}

		c.log.Info("sent acks", zap.Int64s("message-ids", buff))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if len(buff) > 0 {
				sendAcks(ctx)
			}
		case msgID := <-c.ackSendChan:
			buff = append(buff, msgID)
			if len(buff) == ackMaxBatchSize {
				sendAcks(ctx)
				timer.Reset(ackForcedSendTimeout)
			}
		}
	}
}

func (c *Client) ackOutcomingRPC(ctx context.Context, req request) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// set ack callback for our request
	c.ackMux.Lock()
	c.ack[req.ID] = cancel
	c.ackMux.Unlock()

	defer func() {
		c.ackMux.Lock()
		delete(c.ack, req.ID)
		c.ackMux.Unlock()
	}()

	const (
		ackMaxRequestResendRetries = 5
		ackRequestResendTimeout    = time.Second * 15
	)

	retries := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(ackRequestResendTimeout):
			if err := c.write(ctx, req.ID, req.Sequence, req.Input); err != nil {
				c.log.Error("ack timeout resend request", zap.Error(err))
				return
			}

			retries++
			if retries == ackMaxRequestResendRetries {
				c.log.Error("ack retry limit reached", zap.Int64("request-id", req.ID))
				return
			}
		}
	}
}
