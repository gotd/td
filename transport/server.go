package transport

import (
	"context"
	"net"
	"sync/atomic"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/proto/codec"
)

// NewCustomServer creates new MTProto server with custom transport codec.
func NewCustomServer(c Codec, listener net.Listener) *Server {
	return &Server{
		codec:    c,
		listener: listener,
	}
}

// NewIntermediateServer creates new MTProto server with
// Intermediate transport codec.
func NewIntermediateServer(listener net.Listener) *Server {
	return NewCustomServer(codec.Intermediate{}, listener)
}

// Handler is MTProto server connection handler.
type Handler func(ctx context.Context, conn Conn) error

// Server is a simple MTProto server.
type Server struct {
	codec    Codec
	listener net.Listener

	ctx    context.Context
	cancel context.CancelFunc
	closed int64
}

func (s *Server) serveConn(ctx context.Context, handler Handler, c net.Conn) error {
	if err := s.codec.ReadHeader(c); err != nil {
		return xerrors.Errorf("read header: %w", err)
	}

	return handler(ctx, &connection{
		conn:  c,
		codec: s.codec,
	})
}

// Addr returns server address.
func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

// Serve runs server using given listener.
func (s *Server) Serve(ctx context.Context, handler Handler) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		if atomic.LoadInt64(&s.closed) == 1 {
			break
		}
		go func() {
			_ = s.serveConn(s.ctx, handler, conn)
		}()
	}

	return nil
}

// Close stops server and closes given listener.
func (s *Server) Close() error {
	atomic.StoreInt64(&s.closed, 1)

	if s.cancel != nil {
		s.cancel()
	}

	return s.listener.Close()
}
