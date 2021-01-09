package tdsync

import "sync"

// Ready is simple signal primitive which sends signal once.
// This is not allowed to use zero value.
type Ready struct {
	wait chan struct{}
	once sync.Once
}

// NewReady creates new Ready.
func NewReady() *Ready {
	r := &Ready{}
	r.reset()
	return r
}

func (r *Ready) reset() {
	r.wait = make(chan struct{})
	r.once = sync.Once{}
}

// Signal sends ready signal.
// Can be called multiple times.
func (r *Ready) Signal() {
	r.once.Do(func() {
		close(r.wait)
	})
}

// Ready returns waiting channel.
func (r *Ready) Ready() <-chan struct{} {
	return r.wait
}
