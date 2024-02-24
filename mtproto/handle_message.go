package mtproto

import (
	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto"
)

func (c *Conn) handleMessage(msgID int64, b *bin.Buffer) error {
	id, err := b.PeekID()
	if err != nil {
		// Empty body.
		return errors.Wrap(err, "peek message type")
	}

	c.logWithBuffer(b).Debug("Handle message", zap.Int64("msg_id", msgID))
	switch id {
	case mt.NewSessionCreatedTypeID:
		return c.handleSessionCreated(b)
	case mt.BadMsgNotificationTypeID, mt.BadServerSaltTypeID:
		return c.handleBadMsg(b)
	case mt.FutureSaltsTypeID:
		return c.handleFutureSalts(b)
	case proto.MessageContainerTypeID:
		return c.handleContainer(msgID, b)
	case proto.ResultTypeID:
		return c.handleResult(b)
	case mt.PongTypeID:
		return c.handlePong(b)
	case mt.MsgsAckTypeID:
		return c.handleAck(b)
	case proto.GZIPTypeID:
		return c.handleGZIP(msgID, b)
	case mt.MsgDetailedInfoTypeID,
		mt.MsgNewDetailedInfoTypeID:
		return nil
	default:
		return c.handler.OnMessage(b)
	}
}
