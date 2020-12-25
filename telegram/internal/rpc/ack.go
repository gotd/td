package rpc

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// NotifyAcks notifies engine about received acknowledgements.
func (e *Engine) NotifyAcks(ids []int64) {
	for _, id := range ids {
		e.mux.Lock()
		cb, ok := e.ack[id]
		e.mux.Unlock()

		if !ok {
			e.log.Warn("ack callback not set", zap.Int64("msg_id", id))
			continue
		}

		cb()
	}
}

// waitAck blocks until acknowledgement on message id is received.
func (e *Engine) waitAck(ctx context.Context, id int64) error {
	log := e.log.With(zap.Int64("msg_id", id))
	log.Debug("Waiting for Ack")

	done := make(chan struct{})
	var ackOnce sync.Once

	e.mux.Lock()
	e.ack[id] = func() {
		ackOnce.Do(func() {
			close(done)
		})
	}
	e.mux.Unlock()

	defer func() {
		e.mux.Lock()
		delete(e.ack, id)
		e.mux.Unlock()
	}()

	select {
	case <-ctx.Done():
		log.Debug("Ack context done")
		return ctx.Err()
	case <-done:
		log.Debug("Ack received")
		return nil
	}
}
