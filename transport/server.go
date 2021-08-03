package transport

import (
	"context"
	"net"
	"sync"

	"github.com/gotd/td/internal/proto/codec"
	"github.com/gotd/td/internal/tdsync"
)

// NewCustomServer creates new MTProto server with custom transport codec.
// Parameter codec may be nil, it means that MTProto transport will be detected automatically.
func NewCustomServer(c func() Codec, listener net.Listener) *Server {
	return &Server{
		codec:    c,
		listener: &onceCloseListener{Listener: listener},
	}
}

// NewFullServer creates new MTProto server with
// Full transport codec.
func NewFullServer(listener net.Listener) *Server {
	return NewCustomServer(func() Codec {
		return &codec.Full{}
	}, listener)
}

// NewIntermediateServer creates new MTProto server with
// Intermediate transport codec.
func NewIntermediateServer(listener net.Listener) *Server {
	return NewCustomServer(func() Codec {
		return &codec.Intermediate{}
	}, listener)
}

// Handler is MTProto server connection handler.
type Handler func(ctx context.Context, conn Conn) error

// Server is a simple MTProto server.
type Server struct {
	codec    func() Codec
	listener net.Listener

	serveMux sync.Mutex
	serve    *tdsync.CancellableGroup
}

// Addr returns server address.
func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

// Close stops server and closes given listener.
func (s *Server) Close() error {
	s.serveMux.Lock()
	if s.serve != nil {
		s.serve.Cancel()
	}
	s.serveMux.Unlock()

	return s.listener.Close()
}
