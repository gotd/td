package transport

import (
	"context"
	"io"
	"net"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/tdsync"
)

type wrappedConn struct {
	reader io.Reader
	net.Conn
}

func (w wrappedConn) Read(b []byte) (int, error) {
	return w.reader.Read(b)
}

func (s *Server) serveConn(ctx context.Context, handler Handler, conn net.Conn) error {
	if v, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(v); err != nil {
			return xerrors.Errorf("set deadline: %w", err)
		}
	}

	if s.codec == nil {
		transportCodec, reader, err := detectCodec(conn)
		if err != nil {
			return xerrors.Errorf("detect codec: %w", err)
		}

		return handler(ctx, &connection{
			conn: wrappedConn{
				reader: reader,
				Conn:   conn,
			},
			codec: transportCodec,
		})
	}

	transportCodec := s.codec()
	if err := transportCodec.ReadHeader(conn); err != nil {
		return xerrors.Errorf("read header: %w", err)
	}
	return handler(ctx, &connection{
		conn:  conn,
		codec: transportCodec,
	})
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
