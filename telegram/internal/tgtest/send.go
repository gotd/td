package tgtest

import (
	"errors"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
)

func (s *Server) Send(k crypto.AuthKey, encoder bin.Encoder) error {
	conn := s.conns.get(k)
	if conn == nil {
		return errors.New("invalid key: connection not found")
	}

	var buf bin.Buffer
	msg, err := s.encryptData(k, encoder)
	if err != nil {
		return xerrors.Errorf("failed to encrypt message: %w", err)
	}

	if err := msg.Encode(&buf); err != nil {
		return xerrors.Errorf("failed to encode message: %w", err)
	}

	return proto.WriteIntermediate(conn, &buf)
}

func (s *Server) SendResult(k crypto.AuthKey, id int64, msg bin.Encoder) error {
	var buf bin.Buffer

	if err := msg.Encode(&buf); err != nil {
		return xerrors.Errorf("failed to encode result data: %w", err)
	}

	return s.Send(k, &proto.Result{
		RequestMessageID: id,
		Result:           buf.Raw(),
	})
}
