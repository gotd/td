package tdsync

import (
	"sync/atomic"
)

// Ready is simple signal primitive which sends signal once.
// This is not allowed to use zero value.
type Ready struct {
	wait chan struct{}
	done int32
}

// NewReady creates new Ready.
func NewReady() *Ready {
	return &Ready{
		wait: make(chan struct{}),
	}
}

func (r *Ready) reset() {
	r.wait = make(chan struct{})
	atomic.StoreInt32(&r.done, 0)
}

// Signal sends ready signal.
// Can be called multiple times.
func (r *Ready) Signal() {
	if atomic.CompareAndSwapInt32(&r.done, 0, 1) {
		close(r.wait)
	}
}

// Ready returns waiting channel.
func (r *Ready) Ready() <-chan struct{} {
	return r.wait
}
