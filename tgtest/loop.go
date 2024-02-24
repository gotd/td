package tgtest

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/proto/codec"
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
		return errors.Wrap(err, "send")
	}
	return nil
}

func (s *Server) serveConn(ctx context.Context, conn transport.Conn) error {
	s.log.Debug("User connected")
	defer func() {
		_ = conn.Close()
		s.log.Debug("User disconnected")
	}()

	b := new(bin.Buffer)
	for {
		if err := s.read(ctx, conn, b); err != nil {
			return errors.Wrap(err, "read")
		}

		var authKeyID [8]byte
		if err := b.PeekN(authKeyID[:], len(authKeyID)); err != nil {
			return errors.Wrap(err, "peek id")
		}

		// TODO(tdakkota): dispatch by type ID instead?
		if _, ok := s.users.getSession(authKeyID); ok {
			if err := s.rpcHandle(ctx, conn, b); err != nil {
				return errors.Wrap(err, "handle")
			}
			continue
		}

		// If authKeyID not found and is not zero, so send protocol error.
		if authKeyID != [8]byte{} {
			if err := s.sendProtoError(ctx, conn, codec.CodeAuthKeyNotFound); err != nil {
				return errors.Wrap(err, "send AuthKeyNotFound")
			}
			continue
		}

		s.log.Debug("Starting key exchange")
		c := newBufferedConn(conn)
		c.Push(b)

		key, err := s.exchange(ctx, exchangeConn{Conn: c})
		if err != nil {
			var exchangeErr *exchange.ServerExchangeError
			if errors.As(err, &exchangeErr) {
				code := exchangeErr.Code
				if err := s.sendProtoError(ctx, c, code); err != nil {
					return errors.Wrapf(err, "send proto error %v", code)
				}
				return nil
			}
			return errors.Wrap(err, "key exchange failed")
		}

		s.users.addSession(key)
	}
}
