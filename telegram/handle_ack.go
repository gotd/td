package telegram

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
)

func (c *Client) handleAck(b *bin.Buffer) error {
	var ack mt.MsgsAck
	if err := ack.Decode(b); err != nil {
		return xerrors.Errorf("decode: %x", err)
	}
	c.log.With(zap.Int64s("messages", ack.MsgIds)).Debug("Ack")

	c.ackMux.Lock()
	defer c.ackMux.Unlock()

	for _, msgID := range ack.MsgIds {
		fn, found := c.ack[msgID]
		if !found {
			c.log.Warn("ack callback is not set", zap.Int64("message-id", msgID))
			continue
		}

		fn()
		delete(c.ack, msgID)
	}

	return nil
}
