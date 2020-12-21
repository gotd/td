package telegram

import "sync"

type condOnce struct {
	ch        chan struct{}
	closeOnce sync.Once
}

func createCondOnce() *condOnce {
	c := &condOnce{}
	c.Reset()
	return c
}

func (c *condOnce) Reset() {
	// Reset state.
	c.ch = make(chan struct{}, 1)
	c.ch <- struct{}{}
	c.closeOnce = sync.Once{}
}

func (c *condOnce) Done() {
	c.closeOnce.Do(func() {
		// Broadcast all waiters to unlock.
		close(c.ch)
	})
}

func (c *condOnce) WaitIfNeeded() {
	select {
	// Do not block on first call, or on close.
	case <-c.ch:
	default:
		// Channel is not closed and is empty, wait.
		<-c.ch
	}
}
