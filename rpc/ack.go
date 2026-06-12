package rpc

import (
	"context"

	"github.com/gotd/log"
)

// NotifyAcks notifies engine about received acknowledgements.
func (e *Engine) NotifyAcks(ids []int64) {
	e.mux.Lock()
	defer e.mux.Unlock()

	for _, id := range ids {
		ch, ok := e.ack[id]
		if !ok {
			e.log.Debug(context.Background(), "Acknowledge callback not set", log.Int64("msg_id", id))
			continue
		}

		close(ch)
		delete(e.ack, id)
	}
}

func (e *Engine) waitAck(id int64) chan struct{} {
	e.mux.Lock()
	defer e.mux.Unlock()

	ctx := context.Background()
	logger := e.log.With(log.Int64("ack_id", id))
	if c, found := e.ack[id]; found {
		logger.Warn(ctx, "Ack already registered")
		return c
	}

	logger.Debug(ctx, "Waiting for acknowledge")
	c := make(chan struct{})
	e.ack[id] = c
	return c
}

func (e *Engine) removeAck(id int64) {
	e.mux.Lock()
	defer e.mux.Unlock()

	delete(e.ack, id)
}
