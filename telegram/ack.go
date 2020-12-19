package telegram

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/gotd/td/internal/mt"
)

type ackConfig struct {
	// Determines the maximum number of ACKs in a buffer.
	// When buffer reaches this limit, ACKs will be sent immediately.
	MaxBatchSize int

	// The period after which the ACKs will be forcibly sent.
	// (if there is at least one ack in the buffer)
	ForcedSendTimeout time.Duration

	// The maximum number of attempts to send a request before receiving a ACK.
	// If this limit is exceeded, the request fails.
	MaxRequestResendRetries int

	// The period after which the request will be sent again if the ACK has not arrived.
	RequestResendTimeout time.Duration
}

func (cfg *ackConfig) defaultize() {
	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = 20
	}
	if cfg.ForcedSendTimeout.Nanoseconds() <= 0 {
		cfg.ForcedSendTimeout = time.Second * 15
	}
	if cfg.MaxRequestResendRetries <= 0 {
		cfg.MaxRequestResendRetries = 5
	}
	if cfg.RequestResendTimeout.Nanoseconds() <= 0 {
		cfg.RequestResendTimeout = time.Second * 15
	}
}

type acker struct {
	callbacks map[int64]func()
	mux       sync.Mutex

	sendChan chan int64

	client *Client
	log    *zap.Logger
	cfg    ackConfig
}

func newAcker(client *Client, log *zap.Logger, cfg ackConfig) *acker {
	cfg.defaultize()

	return &acker{
		callbacks: map[int64]func(){},
		sendChan:  make(chan int64),
		client:    client,
		log:       log,
		cfg:       cfg,
	}
}

// run starts ack send loop.
func (a *acker) run(ctx context.Context) {
	a.client.wg.Add(1)
	defer a.client.wg.Done()

	var (
		buf = make([]int64, 0, a.cfg.MaxBatchSize)

		// TODO(ernado): remove side-effect.
		timer = time.NewTimer(a.cfg.ForcedSendTimeout)
	)
	defer timer.Stop()

	send := func() {
		defer func() { buf = buf[:0] }()

		if err := a.client.writeServiceMessage(ctx, &mt.MsgsAck{MsgIds: buf}); err != nil {
			a.log.Error("Failed to ACK", zap.Error(err))
			return
		}

		a.log.Info("ACK", zap.Int64s("message_ids", buf))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if len(buf) > 0 {
				send()
			}
		case msgID := <-a.sendChan:
			buf = append(buf, msgID)
			if len(buf) == a.cfg.MaxBatchSize {
				send()
				timer.Reset(a.cfg.ForcedSendTimeout)
			}
		}
	}
}

// rpcRetryUntilAck resends the request if, after a certain time
// (specified in the config), ACK has not been received from the server.
//
// The number of sending attempts is limited (see config).
func (a *acker) rpcRetryUntilAck(ctx context.Context, req request) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set ack callback for request.
	a.mux.Lock()
	a.callbacks[req.ID] = cancel
	a.mux.Unlock()

	defer func() {
		a.mux.Lock()
		delete(a.callbacks, req.ID)
		a.mux.Unlock()
	}()

	retries := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(a.cfg.RequestResendTimeout):
			if err := a.client.write(ctx, req.ID, req.Sequence, req.Input); err != nil {
				a.log.Error("ACK timeout resend request", zap.Error(err))
				return
			}

			retries++
			if retries == a.cfg.MaxRequestResendRetries {
				a.log.Error("ACK retry limit reached", zap.Int64("request_id", req.ID))
				return
			}
		}
	}
}

// handleAcks handles ACKs received from the server.
func (a *acker) handleAcks(msgIDs []int64) {
	a.mux.Lock()
	defer a.mux.Unlock()

	for _, msgID := range msgIDs {
		cb, found := a.callbacks[msgID]
		if !found {
			a.log.Warn("ACK callback is not set", zap.Int64("message_id", msgID))
			continue
		}

		cb()
		delete(a.callbacks, msgID)
	}
}

// sendAck sends ack with provided messageID to the server.
func (a *acker) sendAck(msgID int64) {
	a.sendChan <- msgID
}
