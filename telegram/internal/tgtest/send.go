package tgtest

import (
	"errors"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
)

func (s *Server) Send(k Session, encoder bin.Encoder) error {
	conn := s.conns.get(k)
	if conn == nil {
		return errors.New("invalid key: connection not found")
	}

	var b bin.Buffer
	if err := encoder.Encode(&b); err != nil {
		return xerrors.Errorf("failed to encode data: %w", err)
	}

	data := crypto.EncryptedMessageData{
		SessionID:              k.SessionID,
		MessageDataLen:         int32(b.Len()),
		MessageDataWithPadding: b.Copy(),
	}

	err := s.cipher.EncryptDataTo(k.Key, data, &b)
	if err != nil {
		return xerrors.Errorf("failed to encrypt message: %w", err)
	}

	return conn.Send(s.ctx, &b)
}

func (s *Server) SendResult(k Session, id int64, msg bin.Encoder) error {
	var buf bin.Buffer

	if err := msg.Encode(&buf); err != nil {
		return xerrors.Errorf("failed to encode result data: %w", err)
	}

	return s.Send(k, &proto.Result{
		RequestMessageID: id,
		Result:           buf.Raw(),
	})
}
