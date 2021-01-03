package mtproto

import "sync"

type condOnce struct {
	ch        chan struct{}
	mux       sync.RWMutex
	closeOnce sync.Once
}

func createCondOnce() *condOnce {
	c := &condOnce{}
	c.Reset()
	return c
}

func (c *condOnce) Reset() {
	c.mux.Lock()
	defer c.mux.Unlock()
	// Reset state.
	c.ch = make(chan struct{}, 1)
	c.ch <- struct{}{}
	c.closeOnce = sync.Once{}
}

func (c *condOnce) Done() {
	c.mux.RLock()
	defer c.mux.RUnlock()

	c.closeOnce.Do(func() {
		// Broadcast all waiters to unlock.
		close(c.ch)
	})
}

func (c *condOnce) WaitIfNeeded() {
	c.mux.RLock()
	defer c.mux.RUnlock()

	select {
	// Do not block on first call, or on close.
	case <-c.ch:
	default:
		// Channel is not closed and is empty, wait.
		<-c.ch
	}
}
