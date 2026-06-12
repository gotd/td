package mtproto

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
)

func (c *Conn) handleAck(b *bin.Buffer) error {
	var ack mt.MsgsAck
	if err := ack.Decode(b); err != nil {
		return errors.Wrap(err, "decode")
	}

	c.log.Debug(context.Background(), "Received ack", log.Any("msg_ids", ack.MsgIDs))
	c.rpc.NotifyAcks(ack.MsgIDs)

	return nil
}
