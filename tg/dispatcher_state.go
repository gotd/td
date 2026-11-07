package tg

import "sync"

// dispatcherState is shared by all copies of UpdateDispatcher.
//
// UpdateDispatcher is commonly passed by value (e.g. as telegram.UpdateHandler),
// so every registration must be visible through every copy, and handlers can
// be registered while updates are being dispatched.
type dispatcherState struct {
	mux      sync.RWMutex
	handlers map[uint32]Handler
	fallback Handler
	before   BeforeHandler
}

func newDispatcherState() *dispatcherState {
	return &dispatcherState{
		handlers: map[uint32]Handler{},
	}
}

func (s *dispatcherState) set(typeID uint32, h Handler) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.handlers[typeID] = h
}

func (s *dispatcherState) setFallback(h Handler) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.fallback = h
}

func (s *dispatcherState) setBefore(h BeforeHandler) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.before = h
}

// handler returns before hook (or nil) and handler for given type ID, or
// fallback handler, or nil.
//
// A nil state (zero UpdateDispatcher value) has no handlers.
func (s *dispatcherState) handler(typeID uint32) (BeforeHandler, Handler) {
	if s == nil {
		return nil, nil
	}
	s.mux.RLock()
	defer s.mux.RUnlock()
	if h, ok := s.handlers[typeID]; ok {
		return s.before, h
	}
	return s.before, s.fallback
}
