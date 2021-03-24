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

	c.log.Debug("Received ack", zap.Int64s("msg_ids", ack.MsgIDs))

	var reqIDs []int64
	c.reqMux.Lock()
	for _, id := range ack.MsgIDs {
		if reqID, ok := c.msgToReq[id]; ok {
			reqIDs = append(reqIDs, reqID)
		}
	}
	c.reqMux.Unlock()
	c.rpc.NotifyAcks(reqIDs)

	return nil
}
