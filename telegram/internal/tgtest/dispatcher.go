package tgtest

import (
	"sync"

	"golang.org/x/xerrors"
)

// Dispatcher is a plain handler to map requests by ID.
type Dispatcher struct {
	reqs     map[uint32]Handler
	mux      sync.Mutex
	fallback Handler
}

// NewDispatcher creates new Dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		reqs: map[uint32]Handler{},
	}
}

// OnMessage implements Handler
func (d *Dispatcher) OnMessage(server *Server, req *Request) error {
	id, err := req.Buf.PeekID()
	if err != nil {
		return xerrors.Errorf("peer id: %w", err)
	}

	d.mux.Lock()
	h, ok := d.reqs[id]
	fallback := d.fallback
	d.mux.Unlock()
	if ok {
		return h.OnMessage(server, req)
	}

	if fallback != nil {
		return fallback.OnMessage(server, req)
	}

	return xerrors.Errorf("unexpected type %d", id)
}

// Handle sets handler for given TypeID.
func (d *Dispatcher) Handle(id uint32, h Handler) *Dispatcher {
	d.mux.Lock()
	d.reqs[id] = h
	d.mux.Unlock()
	return d
}

// HandleFunc handler for given TypeID.
func (d *Dispatcher) HandleFunc(id uint32, h func(server *Server, req *Request) error) *Dispatcher {
	return d.Handle(id, HandlerFunc(h))
}

// Fallback sets fallback handler.
func (d *Dispatcher) Fallback(h Handler) *Dispatcher {
	d.mux.Lock()
	d.fallback = h
	d.mux.Unlock()
	return d
}
