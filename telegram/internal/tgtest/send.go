package tgtest

import (
	"context"
	"math"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/tg"
)

func (s *Server) Send(ctx context.Context, k Session, t proto.MessageType, encoder bin.Encoder) error {
	conn, ok := s.users.getConnection(k.AuthKey)
	if !ok {
		return xerrors.Errorf("send %T: invalid key: connection not found", encoder)
	}

	var b bin.Buffer
	if err := encoder.Encode(&b); err != nil {
		return xerrors.Errorf("failed to encode data: %w", err)
	}

	data := crypto.EncryptedMessageData{
		SessionID:              k.SessionID,
		MessageDataLen:         int32(b.Len()),
		MessageDataWithPadding: b.Copy(),
		MessageID:              s.msgID.New(t),
	}

	err := s.cipher.Encrypt(k.AuthKey, data, &b)
	if err != nil {
		return xerrors.Errorf("failed to encrypt message: %w", err)
	}

	return conn.Send(ctx, &b)
}

func (s *Server) sendReq(req *Request, t proto.MessageType, encoder bin.Encoder) error {
	return s.Send(req.RequestCtx, req.Session, t, encoder)
}

func (s *Server) SendResult(req *Request, msg bin.Encoder) error {
	var buf bin.Buffer

	if err := msg.Encode(&buf); err != nil {
		return xerrors.Errorf("failed to encode result data: %w", err)
	}

	if err := s.sendReq(req, proto.MessageServerResponse, &proto.Result{
		RequestMessageID: req.MsgID,
		Result:           buf.Raw(),
	}); err != nil {
		return xerrors.Errorf("send result [%T]: %w", msg, err)
	}

	return nil
}

func (s *Server) SendVector(req *Request, msgs ...bin.Encoder) error {
	return s.SendResult(req, &genericVector{Elems: msgs})
}

func (s *Server) sendSessionCreated(ctx context.Context, k Session, serverSalt int64) error {
	if err := s.Send(ctx, k, proto.MessageFromServer, &mt.NewSessionCreated{
		ServerSalt: serverSalt,
	}); err != nil {
		return xerrors.Errorf("send sessionCreated: %w", err)
	}

	return nil
}

func (s *Server) SendConfig(req *Request) error {
	s.log.Debug("SendConfig")
	return s.SendResult(req, &tg.Config{})
}

func (s *Server) SendPong(req *Request, pingID int64) error {
	if err := s.sendReq(req, proto.MessageServerResponse, &mt.Pong{
		MsgID:  req.MsgID,
		PingID: pingID,
	}); err != nil {
		return xerrors.Errorf("send pong: %w", err)
	}

	return nil
}

func (s *Server) SendEternalSalt(req *Request) error {
	return s.SendFutureSalts(req, mt.FutureSalt{
		ValidSince: 1,
		ValidUntil: math.MaxInt32,
		Salt:       10,
	})
}

func (s *Server) SendFutureSalts(req *Request, salts ...mt.FutureSalt) error {
	if err := s.Send(req.RequestCtx, req.Session, proto.MessageServerResponse, &mt.FutureSalts{
		ReqMsgID: req.MsgID,
		Now:      int(s.clock.Now().Unix()),
		Salts:    salts,
	}); err != nil {
		return xerrors.Errorf("send future salts: %w", err)
	}

	return nil
}

func (s *Server) SendUpdates(ctx context.Context, k Session, updates ...tg.UpdateClass) error {
	if len(updates) == 0 {
		return nil
	}

	if err := s.Send(ctx, k, proto.MessageFromServer, &tg.Updates{
		Updates: updates,
		Date:    int(s.clock.Now().Unix()),
	}); err != nil {
		return xerrors.Errorf("send updates: %w", err)
	}

	return nil
}

func (s *Server) SendAck(ctx context.Context, k Session, ids ...int64) error {
	if err := s.Send(ctx, k, proto.MessageFromServer, &mt.MsgsAck{MsgIDs: ids}); err != nil {
		return xerrors.Errorf("send ack: %w", err)
	}

	return nil
}

func (s *Server) ForceDisconnect(k Session) {
	s.users.deleteConnection(k.AuthKey)
}
