package tdsync

import "sync"

// SinglePerformer helps to control the execution of some operations,
// the execution of which must occur strictly by one participant.
type SinglePerformer struct {
	performing bool
	callbacks  []func()
	mux        sync.Mutex
}

// NewSinglePerformer creates new SinglePerformer.
func NewSinglePerformer() *SinglePerformer {
	return new(SinglePerformer)
}

// Try tries to obtain permission to perform an operation.
// If function returned true - the caller can perform operation.
// After operation was performed, the callback function must be called.
//
// Otherwise if function returned false - some other caller already performing an operation.
// You must call callback function, which waits when operation was performed.
func (p *SinglePerformer) Try() (func(), bool) {
	p.mux.Lock()
	defer p.mux.Unlock()

	if !p.performing {
		p.performing = true
		return p.onPerformed, true
	}

	return p.waiter(), false
}

func (p *SinglePerformer) waiter() func() {
	c := make(chan struct{})
	p.callbacks = append(p.callbacks, func() { close(c) })

	return func() {
		<-c
	}
}

func (p *SinglePerformer) onPerformed() {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.performing = false
	for _, cb := range p.callbacks {
		cb()
	}

	p.callbacks = nil
}
