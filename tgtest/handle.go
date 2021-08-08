package tgtest

import (
	"context"
	"encoding/binary"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/tgerr"
)

func (s *Server) rpcHandle(ctx context.Context, conn *connection) error {
	var (
		session Session
		b       bin.Buffer
	)

	defer func() {
		s.users.deleteConnection(session.ID)
	}()
	for {
		if err := s.read(ctx, conn, &b); err != nil {
			return xerrors.Errorf("read: %w", err)
		}

		m := &crypto.EncryptedMessage{}
		if err := m.DecodeWithoutCopy(&b); err != nil {
			return xerrors.Errorf("decode encrypted message: %w", err)
		}

		key, ok := s.users.getSession(m.AuthKeyID)
		if !ok {
			return xerrors.New("invalid session")
		}

		msg, err := s.cipher.Decrypt(key, m)
		if err != nil {
			return xerrors.Errorf("decrypt message: %w", err)
		}

		session = Session{
			ID:      msg.SessionID,
			AuthKey: key,
		}
		if !conn.didSentCreated() {
			s.users.addConnection(msg.SessionID, conn)

			s.log.Debug("Send handleSessionCreated event", zap.Inline(session))
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
		zap.Inline(req.Session),
		zap.Int64("msg_id", req.MsgID),
		zap.String("type", s.types.Get(id)),
	)

	// TODO(tdakkota): unpack all containers
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

	case proto.GZIPTypeID:
		var content proto.GZIP
		if err := content.Decode(in); err != nil {
			return xerrors.Errorf("gzip: %w", err)
		}
		req.Buf = &bin.Buffer{Buf: content.Data}

	case proto.MessageContainerTypeID:
		var container proto.MessageContainer
		if err := container.Decode(in); err != nil {
			return xerrors.Errorf("container: %w", err)
		}

		var err error
		for _, msg := range container.Messages {
			err = multierr.Append(err, s.handle(&Request{
				DC:         req.DC,
				Session:    req.Session,
				MsgID:      msg.ID,
				Buf:        &bin.Buffer{Buf: msg.Body},
				RequestCtx: req.RequestCtx,
			}))
		}
		return err
	}

	if err := s.handler.OnMessage(s, req); err != nil {
		var rpcErr *tgerr.Error
		if xerrors.As(err, &rpcErr) {
			return s.SendErr(req, rpcErr)
		}
		return err
	}

	return nil
}
