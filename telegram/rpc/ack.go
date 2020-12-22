package rpc

import (
	"context"

	"go.uber.org/zap"
)

// NotifyACKs notifies engine about received ACKs.
func (e *Engine) NotifyACKs(msgIDs []int64) {
	e.ackMux.RLock()
	defer e.ackMux.RUnlock()

	for _, msgID := range msgIDs {
		cb, ok := e.ack[msgID]
		if !ok {
			e.log.Warn("ack callback not set", zap.Int64("msg_id", msgID))
			continue
		}

		cb()
	}
}

// waitACK waits to receive ACK from the server until the context is canceled.
// If ACK was received - returns nil.
// If the context was canceled before the ACK was received, it returns a context error.
func (e *Engine) waitACK(ctx context.Context, msgID int64) error {
	got := make(chan struct{})

	e.ackMux.Lock()
	e.ack[msgID] = func() { close(got) }
	e.ackMux.Unlock()

	defer func() {
		e.ackMux.Lock()
		delete(e.ack, msgID)
		e.ackMux.Unlock()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-got:
		return nil
	}
}
