package tgtest

import (
	"context"
	"encoding/binary"
	"encoding/hex"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/transport"
)

type Session struct {
	SessionID int64
	crypto.AuthKey
}

func (s *Server) rpcHandle(ctx context.Context, conn *connection) error {
	var b bin.Buffer
	for {
		b.Reset()
		if err := conn.Recv(ctx, &b); err != nil {
			return err
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
			if err := s.sendSessionCreated(ctx, session, salt); err != nil {
				return err
			}
			conn.sentCreated()
		}

		// Buffer now contains plaintext message payload.
		b.ResetTo(msg.Data())

		if err := s.handle(&Request{
			DC:         s.dcID,
			Session:    session,
			MsgID:      msg.MessageID,
			Buf:        &b,
			RequestCtx: ctx,
		}); err != nil {
			return xerrors.Errorf("handle: %w", err)
		}
	}
}

func (s *Server) handle(req *Request) error {
	in := req.Buf
	id, err := in.PeekID()
	if err != nil {
		return xerrors.Errorf("peek id: %w", err)
	}

	s.log.Debug("Got request",
		zap.String("key_id", hex.EncodeToString(req.Session.ID[:])),
		zap.Int64("msg_id", req.MsgID),
		zap.String("type", s.types.Get(id)),
	)

	switch id {
	case mt.PingDelayDisconnectRequestTypeID:
		pingReq := mt.PingDelayDisconnectRequest{}
		if err := pingReq.Decode(in); err != nil {
			return err
		}

		return s.SendPong(req, pingReq.PingID)
	case mt.PingRequestTypeID:
		pingReq := mt.PingRequest{}
		if err := pingReq.Decode(in); err != nil {
			return err
		}

		return s.SendPong(req, pingReq.PingID)

	case mt.GetFutureSaltsRequestTypeID:
		saltsRequest := mt.GetFutureSaltsRequest{}
		if err := saltsRequest.Decode(in); err != nil {
			return err
		}

		return s.SendEternalSalt(req)

	case mt.RPCDropAnswerRequestTypeID:
		drop := mt.RPCDropAnswerRequest{}
		if err := drop.Decode(in); err != nil {
			return err
		}

		return s.SendResult(req, &mt.RPCAnswerDropped{
			MsgID: req.MsgID,
		})
	}

	if err := s.dispatcher.OnMessage(s, req); err != nil {
		var rpcErr *tgerr.Error
		if xerrors.As(err, &rpcErr) {
			return s.SendErr(req, rpcErr)
		}
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
