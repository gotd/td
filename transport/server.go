package transport

import (
	"context"
	"net"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/proto/codec"
)

// NewCustomServer creates new MTProto server with custom transport codec.
func NewCustomServer(c func() Codec, listener net.Listener) *Server {
	return &Server{
		codec:    c,
		listener: listener,
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

	ctx    context.Context
	cancel context.CancelFunc
}

func (s *Server) serveConn(ctx context.Context, handler Handler, c net.Conn) error {
	transportCodec := s.codec()
	if err := transportCodec.ReadHeader(c); err != nil {
		return xerrors.Errorf("read header: %w", err)
	}

	return handler(ctx, &connection{
		conn:  c,
		codec: transportCodec,
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
		go func() {
			_ = s.serveConn(s.ctx, handler, conn)
		}()
	}
}

// Close stops server and closes given listener.
func (s *Server) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	return s.listener.Close()
}
