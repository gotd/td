package tgtest

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/proto/codec"
	"github.com/gotd/td/transport"
)

func (s *Server) read(ctx context.Context, conn *connection, b *bin.Buffer) error {
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

func (s *Server) serveConn(ctx context.Context, conn transport.Conn) (err error) {
	s.log.Debug("User connected")
	defer s.log.Debug("User disconnected")

	defer func() {
		_ = conn.Close()
	}()

	b := new(bin.Buffer)
	if err := conn.Recv(ctx, b); err != nil {
		return xerrors.Errorf("new conn read: %w", err)
	}

	var authKeyID [8]byte
	if err := b.PeekN(authKeyID[:], len(authKeyID)); err != nil {
		return xerrors.Errorf("peek id: %w", err)
	}

	c := newBufferedConn(conn)
	c.Push(b)
	conn = c

	// TODO(tdakkota): dispatch by type ID instead?
	if _, ok := s.users.getSession(authKeyID); !ok {
		// If authKeyID not found and is not zero, so drop buffered message and send protocol error.
		if authKeyID != [8]byte{} {
			c.Pop()

			if err := s.sendProtoError(ctx, conn, codec.CodeAuthKeyNotFound); err != nil {
				return xerrors.Errorf("send AuthKeyNotFound: %w", err)
			}
		}

		s.log.Debug("Starting key exchange")
		key, err := s.exchange(ctx, conn)
		if err != nil {
			return xerrors.Errorf("key exchange failed: %w", err)
		}
		s.users.addSession(key)
	} else {
		s.log.Debug("Session already created, skip key exchange")
	}
	wrappedConn := &connection{
		Conn: conn,
	}

	return s.rpcHandle(ctx, wrappedConn)
}
