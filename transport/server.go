package transport

import (
	"context"
	"net"
	"sync"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/proto/codec"
	"github.com/gotd/td/internal/tdsync"
)

// NewCustomServer creates new MTProto server with custom transport codec.
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

func (s *Server) serveConn(ctx context.Context, handler Handler, c net.Conn) error {
	transportCodec := s.codec()
	if err := transportCodec.ReadHeader(c); err != nil {
		return xerrors.Errorf("read header: %w", err)
	}

	if v, ok := ctx.Deadline(); ok {
		if err := c.SetDeadline(v); err != nil {
			return xerrors.Errorf("set deadline: %w", err)
		}
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
func (s *Server) Serve(serveCtx context.Context, handler Handler) error {
	s.serveMux.Lock()
	s.serve = tdsync.NewCancellableGroup(serveCtx)
	s.serveMux.Unlock()

	s.serve.Go(func(ctx context.Context) error {
		<-ctx.Done()
		_ = s.listener.Close()
		return nil
	})
	s.serve.Go(func(ctx context.Context) error {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-serveCtx.Done():
					return serveCtx.Err()
				case <-ctx.Done():
					// If parent context is not done, so
					// serve group context is canceled by Close.
					return nil
				default:
				}
				return err
			}

			s.serve.Go(func(ctx context.Context) error {
				return s.serveConn(ctx, handler, conn)
			})
		}
	})

	return s.serve.Wait()
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
