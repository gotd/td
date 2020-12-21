package telegram

type condOnce struct {
	ch        chan struct{}
}

func createCondOnce() *condOnce {
	c := &condOnce{
	}
	c.Reset()
	return c
}

func (c *condOnce) Reset() {
	// Reset state.
	c.ch = make(chan struct{}, 1)
	c.ch <- struct{}{}
}

func (c *condOnce) Done() {
	// Broadcast all waiters to unlock.
	close(c.ch)
}

func (c *condOnce) WaitIfNeeded() {
	select {
	case _, _ = <-c.ch: // do not block on first call, or on close
	default: // channel is not closed and is empty, wait
		<-c.ch
	}
}
