package mtproto

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
)

func (c *Conn) handleAck(b *bin.Buffer) error {
	var ack mt.MsgsAck
	if err := ack.Decode(b); err != nil {
		return xerrors.Errorf("decode: %w", err)
	}

	c.log.With(zap.Int64s("messages", ack.MsgIds)).Debug("Ack")

	c.rpc.NotifyAcks(ack.MsgIds)

	return nil
}
