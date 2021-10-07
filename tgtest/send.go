package tgtest

import (
	"context"
	"math"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/mt"
	"github.com/nnqq/td/internal/proto"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
)

const (
	// MessageServerResponse is a message type of RPC calls result.
	MessageServerResponse = proto.MessageServerResponse
	// MessageFromServer is a message type of server-side updates.
	MessageFromServer = proto.MessageFromServer
)

// Send sends given message to user session k.
// Parameter t denotes MTProto message type. It should be MessageServerResponse or MessageFromServer.
func (s *Server) Send(ctx context.Context, k Session, t proto.MessageType, message bin.Encoder) error {
	conn, ok := s.users.getConnection(k.ID)
	if !ok {
		return xerrors.Errorf("send %T: invalid key: connection %s not found", message, k.AuthKey.String())
	}

	var b bin.Buffer
	if err := message.Encode(&b); err != nil {
		return xerrors.Errorf("encode: %w", err)
	}

	data := crypto.EncryptedMessageData{
		SessionID:              k.ID,
		MessageDataLen:         int32(b.Len()),
		MessageDataWithPadding: b.Copy(),
		MessageID:              s.msgID.New(t),
	}

	err := s.cipher.Encrypt(k.AuthKey, data, &b)
	if err != nil {
		return xerrors.Errorf("encrypt: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, s.writeTimeout)
	defer cancel()

	if err := conn.Send(ctx, &b); err != nil {
		return xerrors.Errorf("send: %w", err)
	}

	return nil
}

func (s *Server) sendReq(req *Request, t proto.MessageType, encoder bin.Encoder) error {
	return s.Send(req.RequestCtx, req.Session, t, encoder)
}

// SendResult sends RPC answer using msg as result.
func (s *Server) SendResult(req *Request, msg bin.Encoder) error {
	var buf bin.Buffer

	if err := msg.Encode(&buf); err != nil {
		return xerrors.Errorf("encode result: %w", err)
	}

	if err := s.sendReq(req, proto.MessageServerResponse, &proto.Result{
		RequestMessageID: req.MsgID,
		Result:           buf.Raw(),
	}); err != nil {
		return xerrors.Errorf("send result [%T]: %w", msg, err)
	}

	return nil
}

// SendGZIP sends RPC answer and packs it into proto.GZIP.
func (s *Server) SendGZIP(req *Request, msg bin.Encoder) error {
	var buf bin.Buffer

	if err := msg.Encode(&buf); err != nil {
		return xerrors.Errorf("encode gzip data: %w", err)
	}

	return s.SendResult(req, proto.GZIP{Data: buf.Buf})
}

// SendErr sends RPC answer using given error as result.
func (s *Server) SendErr(req *Request, e *tgerr.Error) error {
	return s.SendResult(req, &mt.RPCError{
		ErrorCode:    e.Code,
		ErrorMessage: e.Message,
	})
}

// SendBool sends RPC answer using given bool as result.
// Usually used in methods without explicit response.
func (s *Server) SendBool(req *Request, r bool) error {
	var msg tg.BoolClass = &tg.BoolTrue{}
	if !r {
		msg = &tg.BoolFalse{}
	}
	return s.SendResult(req, msg)
}

// SendVector sends RPC answer using given vector as result.
func (s *Server) SendVector(req *Request, msgs ...bin.Encoder) error {
	return s.SendResult(req, &genericVector{Elems: msgs})
}

// sendSessionCreated sends mt.NewSessionCreated `new_session_created` notification.
func (s *Server) sendSessionCreated(ctx context.Context, k Session, serverSalt int64) error {
	if err := s.Send(ctx, k, proto.MessageFromServer, &mt.NewSessionCreated{
		FirstMsgID: s.msgID.New(proto.MessageFromClient),
		ServerSalt: serverSalt,
	}); err != nil {
		return xerrors.Errorf("send sessionCreated: %w", err)
	}

	return nil
}

// SendPong sends response for mt.PingRequest request.
func (s *Server) SendPong(req *Request, pingID int64) error {
	if err := s.sendReq(req, proto.MessageServerResponse, &mt.Pong{
		MsgID:  req.MsgID,
		PingID: pingID,
	}); err != nil {
		return xerrors.Errorf("send pong: %w", err)
	}

	return nil
}

// SendEternalSalt sends response for mt.GetFutureSaltsRequest.
// It sends an `eternal` salt, which valid until maximum possible date.
func (s *Server) SendEternalSalt(req *Request) error {
	return s.SendFutureSalts(req, mt.FutureSalt{
		ValidSince: 1,
		ValidUntil: math.MaxInt32,
		Salt:       10,
	})
}

// SendFutureSalts sends response for mt.GetFutureSaltsRequest.
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

// SendUpdates sends given updates to user session k.
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

// SendAck sends acknowledgment for received message.
func (s *Server) SendAck(ctx context.Context, k Session, ids ...int64) error {
	if err := s.Send(ctx, k, proto.MessageFromServer, &mt.MsgsAck{MsgIDs: ids}); err != nil {
		return xerrors.Errorf("send ack: %w", err)
	}

	return nil
}

// ForceDisconnect forcibly disconnect user from server.
// It deletes MTProto session (session_id), but not auth key.
func (s *Server) ForceDisconnect(k Session) {
	s.users.deleteConnection(k.ID)
}
