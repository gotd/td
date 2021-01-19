package rpc

import (
	"go.uber.org/zap"
)

// NotifyAcks notifies engine about received acknowledgements.
func (e *Engine) NotifyAcks(ids []int64) {
	e.mux.Lock()
	defer e.mux.Unlock()

	for _, id := range ids {
		ch, ok := e.ack[id]
		if !ok {
			e.log.Debug("Acknowledge callback not set", zap.Int64("msg_id", id))
			continue
		}

		close(ch)
		delete(e.ack, id)
	}
}

func (e *Engine) waitAck(id int64) chan struct{} {
	e.mux.Lock()
	defer e.mux.Unlock()

	log := e.log.With(zap.Int64("ack_id", id))
	if c, found := e.ack[id]; found {
		log.Warn("Ack already registered")
		return c
	}

	log.Debug("Waiting for acknowledge")
	c := make(chan struct{})
	e.ack[id] = c
	return c
}

func (e *Engine) removeAck(id int64) {
	e.mux.Lock()
	defer e.mux.Unlock()

	delete(e.ack, id)
}
