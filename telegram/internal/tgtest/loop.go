package tgtest

import (
	"context"
	"encoding/binary"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/transport"
)

type Session struct {
	SessionID int64
	crypto.AuthKeyWithID
}

func (s *Server) rpcHandle(ctx context.Context, conn connection) error {
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

		session := Session{
			SessionID:     msg.SessionID,
			AuthKeyWithID: key,
		}
		if !conn.sentCreated {
			if err := s.Send(session, &mt.NewSessionCreated{
				ServerSalt: int64(binary.LittleEndian.Uint64(key.AuthKeyID[:])),
			}); err != nil {
				return err
			}
			conn.sentCreated = true
		}

		// Buffer now contains plaintext message payload.
		b.ResetTo(msg.Data())

		if err := s.handler.OnMessage(session, msg.MessageID, &b); err != nil {
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
	wrappedConn := connection{
		Connection: conn,
	}
	s.users.createSession(k, wrappedConn)

	err = s.handler.OnNewClient(k)
	if err != nil {
		return xerrors.Errorf("OnNewClient handler failed: %w", err)
	}

	return s.rpcHandle(ctx, wrappedConn)
}
