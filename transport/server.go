package transport

import (
	"context"
	"net"
	"sync/atomic"
)

func NewCustomServer(codec Codec, listener net.Listener) *Server {
	return &Server{
		codec:    codec,
		listener: listener,
	}
}

type Server struct {
	codec    Codec
	listener net.Listener
	handler  func(ctx context.Context, conn Connection) error

	ctx    context.Context
	cancel context.CancelFunc
	closed int64
}

func (s *Server) serveConn(ctx context.Context, c net.Conn) error {
	if err := s.codec.ReadHeader(c); err != nil {
		return err
	}

	return s.handler(ctx, Connection{
		c, s.codec,
	})
}

func (s *Server) Serve(ctx context.Context) error {
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
			_ = s.serveConn(s.ctx, conn)
		}()
	}

	return nil
}

func (s *Server) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	return s.listener.Close()
}
