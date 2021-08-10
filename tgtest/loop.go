package tgtest

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/proto/codec"
	"github.com/gotd/td/transport"
)

func (s *Server) read(ctx context.Context, conn transport.Conn, b *bin.Buffer) error {
	b.Reset()

	ctx, cancel := context.WithTimeout(ctx, s.readTimeout)
	defer cancel()
	if err := conn.Recv(ctx, b); err != nil {
		return err
	}

	return nil
}

func (s *Server) sendProtoError(ctx context.Context, conn transport.Conn, e int32) error {
	var buf bin.Buffer
	buf.PutInt32(-e)

	ctx, cancel := context.WithTimeout(ctx, s.writeTimeout)
	defer cancel()

	if err := conn.Send(ctx, &buf); err != nil {
		return xerrors.Errorf("send: %w", err)
	}
	return nil
}

func (s *Server) serveConn(ctx context.Context, conn transport.Conn) error {
	s.log.Debug("User connected")
	defer s.log.Debug("User disconnected")

	c := newBufferedConn(conn)
	defer func() {
		_ = c.Close()
	}()

	b := new(bin.Buffer)
	for {
		if err := s.read(ctx, conn, b); err != nil {
			return xerrors.Errorf("read: %w", err)
		}

		var authKeyID [8]byte
		if err := b.PeekN(authKeyID[:], len(authKeyID)); err != nil {
			return xerrors.Errorf("peek id: %w", err)
		}

		// TODO(tdakkota): dispatch by type ID instead?
		if _, ok := s.users.getSession(authKeyID); !ok {
			c.Push(b)

			// If authKeyID not found and is not zero, so drop buffered message and send protocol error.
			if authKeyID != [8]byte{} {
				c.Pop()

				if err := s.sendProtoError(ctx, c, codec.CodeAuthKeyNotFound); err != nil {
					return xerrors.Errorf("send AuthKeyNotFound: %w", err)
				}
			}

			s.log.Debug("Starting key exchange")
			key, err := s.exchange(ctx, c)
			if err != nil {
				return xerrors.Errorf("key exchange failed: %w", err)
			}
			s.users.addSession(key)
			continue
		}

		if err := s.rpcHandle(ctx, &connection{
			Conn: conn,
		}, b); err != nil {
			return xerrors.Errorf("handle: %w", err)
		}
	}
}
