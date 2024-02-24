package mtproto

import (
	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
)

func (c *Conn) handleAck(b *bin.Buffer) error {
	var ack mt.MsgsAck
	if err := ack.Decode(b); err != nil {
		return errors.Wrap(err, "decode")
	}

	c.log.Debug("Received ack", zap.Int64s("msg_ids", ack.MsgIDs))
	c.rpc.NotifyAcks(ack.MsgIDs)

	return nil
}
