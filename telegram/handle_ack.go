package telegram

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/mt"
)

func (c *Client) handleAck(b *bin.Buffer) error {
	var ack mt.MsgsAck
	if err := ack.Decode(b); err != nil {
		return xerrors.Errorf("failed to decode: %x", err)
	}
	c.log.With(zap.Int64s("messages", ack.MsgIds)).Debug("Ack")
	return nil
}
