package mtproto

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/mt"
)

func (c *Conn) handleAck(b *bin.Buffer) error {
	var ack mt.MsgsAck
	if err := ack.Decode(b); err != nil {
		return xerrors.Errorf("decode: %w", err)
	}

	c.log.Debug("Received ack", zap.Int64s("msg_ids", ack.MsgIDs))
	c.rpc.NotifyAcks(ack.MsgIDs)

	return nil
}
