package tgtest

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/transport"
)

func (s *Server) rpcHandle(ctx context.Context, k Session, conn transport.Connection) error {
	var b bin.Buffer
	for {
		b.Reset()
		if err := conn.Recv(ctx, &b); err != nil {
			return xerrors.Errorf("read from client: %w", err)
		}

		msg, err := s.cipher.DecryptDataFrom(k.Key, 0, &b)
		if err != nil {
			return xerrors.Errorf("failed to decrypt: %w", err)
		}
		k.SessionID = msg.SessionID

		// Buffer now contains plaintext message payload.
		b.ResetTo(msg.Data())

		if err := s.handler.OnMessage(k, msg.MessageID, &b); err != nil {
			return xerrors.Errorf("failed to call handler: %w", err)
		}
	}
}

func (s *Server) serveConn(ctx context.Context, conn transport.Connection) error {
	var session Session
	defer func() {
		s.conns.delete(session)
		_ = conn.Close()
	}()

	session, err := s.exchange(ctx, conn)
	if err != nil {
		return xerrors.Errorf("key exchange failed: %w", err)
	}
	s.conns.add(session, conn)

	err = s.handler.OnNewClient(session)
	if err != nil {
		return xerrors.Errorf("OnNewClient handler failed: %w", err)
	}

	return s.rpcHandle(ctx, session, conn)
}
