package tgtest

import (
	"context"

	"github.com/gotd/td/internal/crypto"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/transport"
)

type Session struct {
	SessionID int64
	crypto.AuthKeyWithID
}

func (s *Server) rpcHandle(ctx context.Context, conn transport.Connection) error {
	var b bin.Buffer
	for {
		b.Reset()
		if err := conn.Recv(ctx, &b); err != nil {
			return xerrors.Errorf("read from client: %w", err)
		}

		m := &crypto.EncryptedMessage{}
		if err := m.Decode(&b); err != nil {
			return xerrors.Errorf("encrypted message decode: %w", err)
		}

		key, ok := s.users.getSession(m.AuthKeyID)
		if !ok {
			return xerrors.Errorf("invalid session")
		}

		msg, err := s.cipher.Decrypt(key, m)
		if err != nil {
			return xerrors.Errorf("failed to decrypt: %w", err)
		}

		// Buffer now contains plaintext message payload.
		b.ResetTo(msg.Data())

		if err := s.handler.OnMessage(Session{
			SessionID:     msg.SessionID,
			AuthKeyWithID: key,
		}, msg.MessageID, &b); err != nil {
			return xerrors.Errorf("failed to call handler: %w", err)
		}
	}
}

func (s *Server) serveConn(ctx context.Context, conn transport.Connection) (err error) {
	var k crypto.AuthKeyWithID
	defer func() {
		s.users.deleteConnection(k)
		_ = conn.Close()
	}()

	k, err = s.exchange(ctx, conn)
	if err != nil {
		return xerrors.Errorf("key exchange failed: %w", err)
	}
	s.users.createSession(k, conn)

	err = s.handler.OnNewClient(k)
	if err != nil {
		return xerrors.Errorf("OnNewClient handler failed: %w", err)
	}

	return s.rpcHandle(ctx, conn)
}
