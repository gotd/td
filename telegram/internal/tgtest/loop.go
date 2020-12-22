package tgtest

import (
	"context"
	"encoding/binary"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/transport"
)

type Session struct {
	SessionID int64
	crypto.AuthKeyWithID
}

func (s *Server) rpcHandle(ctx context.Context, conn *connection) error {
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
		if !conn.didSentCreated() {
			salt := int64(binary.LittleEndian.Uint64(key.AuthKeyID[:]))
			if err := s.sendSessionCreated(session, salt); err != nil {
				return err
			}
			conn.sentCreated()
		}

		// Buffer now contains plaintext message payload.
		b.ResetTo(msg.Data())

		if err := s.handler.OnMessage(session, msg.MessageID, &b); err != nil {
			return xerrors.Errorf("failed to call handler: %w", err)
		}
	}
}

func (s *Server) serveConn(ctx context.Context, conn transport.Conn) (err error) {
	var k crypto.AuthKeyWithID
	defer func() {
		s.users.deleteConnection(k)
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

	k, ok := s.users.getSession(authKeyID)
	if !ok {
		k, err = s.exchange(ctx, b, conn)
		if err != nil {
			return xerrors.Errorf("key exchange failed: %w", err)
		}
		s.users.addSession(k)
	}
	wrappedConn := &connection{
		Conn: conn,
	}
	s.users.addConnection(k, wrappedConn)

	return s.rpcHandle(ctx, wrappedConn)
}
