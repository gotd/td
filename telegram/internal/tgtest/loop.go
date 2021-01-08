package tgtest

import (
	"context"
	"encoding/binary"
	"errors"
	"io"

	"github.com/gotd/td/internal/mt"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/transport"
)

type Session struct {
	SessionID int64
	crypto.AuthKey
}

func (s *Server) rpcHandle(ctx context.Context, conn *connection) error {
	var b bin.Buffer
	var key crypto.AuthKey
	for {
		b.Reset()
		if err := conn.Recv(ctx, &b); err != nil {
			if errors.Is(err, io.EOF) {
				// Client disconnected.
				s.users.deleteConnection(key)
				return nil
			}
			return xerrors.Errorf("read from client: %w", err)
		}

		m := &crypto.EncryptedMessage{}
		if err := m.Decode(&b); err != nil {
			return xerrors.Errorf("encrypted message decode: %w", err)
		}

		k, ok := s.users.getSession(m.AuthKeyID)
		if !ok {
			return xerrors.Errorf("invalid session")
		}
		key := k

		msg, err := s.cipher.Decrypt(key, m)
		if err != nil {
			return xerrors.Errorf("failed to decrypt: %w", err)
		}

		session := Session{
			SessionID: msg.SessionID,
			AuthKey:   key,
		}
		if !conn.didSentCreated() {
			s.log.Debug("Send handleSessionCreated event")
			salt := int64(binary.LittleEndian.Uint64(key.ID[:]))
			if err := s.sendSessionCreated(session, salt); err != nil {
				return err
			}
			conn.sentCreated()
		}

		// Buffer now contains plaintext message payload.
		b.ResetTo(msg.Data())

		if err := s.handle(session, msg.MessageID, &b); err != nil {
			return xerrors.Errorf("handle: %w", err)
		}
	}
}

func (s *Server) handle(session Session, msgID int64, in *bin.Buffer) error {
	id, err := in.PeekID()
	if err != nil {
		return xerrors.Errorf("peek id: %w", err)
	}

	switch id {
	case mt.PingDelayDisconnectRequestTypeID:
		pingReq := mt.PingDelayDisconnectRequest{}
		if err := pingReq.Decode(in); err != nil {
			return err
		}

		return s.SendPong(session, msgID, pingReq.PingID)
	case mt.PingRequestTypeID:
		pingReq := mt.PingRequest{}
		if err := pingReq.Decode(in); err != nil {
			return err
		}

		return s.SendPong(session, msgID, pingReq.PingID)
	}

	if err := s.handler.OnMessage(session, msgID, in); err != nil {
		return xerrors.Errorf("failed to call handler: %w", err)
	}

	return nil
}

func (s *Server) serveConn(ctx context.Context, conn transport.Conn) (err error) {
	s.log.Debug("User connected")

	var k crypto.AuthKey
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
		conn := NewBufferedConn(conn)
		conn.Push(b)

		s.log.Debug("Starting key exchange")
		k, err = s.exchange(ctx, conn)
		if err != nil {
			return xerrors.Errorf("key exchange failed: %w", err)
		}
		s.users.addSession(k)
	} else {
		s.log.Debug("Session already created, skip key exchange")
	}
	wrappedConn := &connection{
		Conn: conn,
	}
	s.users.addConnection(k, wrappedConn)

	return s.rpcHandle(ctx, wrappedConn)
}
