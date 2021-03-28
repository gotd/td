package mtproto

import (
	"context"

	"github.com/gotd/td/bin"
)

func (c *Conn) writeContentMessage(ctx context.Context, msgID int64, seqNo int32, message bin.Encoder) error {
	return c.write(ctx, msgID, seqNo, message)
}

func (c *Conn) writeServiceMessage(ctx context.Context, message bin.Encoder) error {
	msgID, seqNo := c.nextMsgSeq(false)
	return c.write(ctx, msgID, seqNo, message)
}

func (c *Conn) write(ctx context.Context, msgID int64, seqNo int32, message bin.Encoder) error {
	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(msgID, seqNo, message, b); err != nil {
		return err
	}

	if err := c.conn.Send(ctx, b); err != nil {
		return err
	}

	return nil
}

func (c *Conn) nextMsgSeq(content bool) (msgID int64, seqNo int32) {
	c.reqMux.Lock()
	defer c.reqMux.Unlock()

	msgID = c.newMessageID()

	// Computing current sequence number (seqno).
	// This should be serialized with new message id generation.
	//
	// See https://github.com/gotd/td/issues/245 for reference.
	seqNo = c.sentContentMessages * 2
	if content {
		seqNo++
		c.sentContentMessages++
	}

	return
}
