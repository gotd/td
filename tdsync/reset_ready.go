package tdsync

import "sync"

// ResetReady is like Ready, but can be Reset.
type ResetReady struct {
	ready Ready
	lock  sync.Mutex
}

// NewResetReady creates new ResetReady.
func NewResetReady() *ResetReady {
	return &ResetReady{
		ready: Ready{
			wait: make(chan struct{}),
		},
	}
}

// Reset resets underlying Ready.
func (r *ResetReady) Reset() {
	r.lock.Lock()
	r.ready.Signal()
	r.ready.reset()
	r.lock.Unlock()
}

// Signal sends ready signal.
// Can be called multiple times.
func (r *ResetReady) Signal() {
	r.lock.Lock()
	r.ready.Signal()
	r.lock.Unlock()
}

// Ready returns waiting channel.
func (r *ResetReady) Ready() <-chan struct{} {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.ready.Ready()
}
