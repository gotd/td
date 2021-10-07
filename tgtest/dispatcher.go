package tgtest

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
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
		return xerrors.Errorf("peek id: %w", err)
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

	return xerrors.Errorf("unexpected type %#x", id)
}

// Handle sets handler for given TypeID.
func (d *Dispatcher) Handle(id uint32, h Handler) *Dispatcher {
	d.mux.Lock()
	d.reqs[id] = h
	d.mux.Unlock()
	return d
}

// HandleFunc sets handler for given TypeID.
func (d *Dispatcher) HandleFunc(id uint32, h func(server *Server, req *Request) error) *Dispatcher {
	return d.Handle(id, HandlerFunc(h))
}

// Result sets constant result for given TypeID.
// NB: it uses rpc_result to pack given encoder.
func (d *Dispatcher) Result(id uint32, msg bin.Encoder) *Dispatcher {
	return d.HandleFunc(id, func(server *Server, req *Request) error {
		return server.SendResult(req, msg)
	})
}

// Vector sets constant Vector result for given TypeID.
// NB: it uses rpc_result to pack generic vector with given encoders.
func (d *Dispatcher) Vector(id uint32, msgs ...bin.Encoder) *Dispatcher {
	return d.HandleFunc(id, func(server *Server, req *Request) error {
		return server.SendVector(req, msgs...)
	})
}

// Fallback sets fallback handler.
func (d *Dispatcher) Fallback(h Handler) *Dispatcher {
	d.mux.Lock()
	d.fallback = h
	d.mux.Unlock()
	return d
}
