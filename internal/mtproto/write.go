package mtproto

import (
	"context"
	"sync/atomic"

	"github.com/gotd/td/bin"
)

func (c *Conn) writeContentMessage(ctx context.Context, id int64, message bin.Encoder) error {
	return c.write(ctx, id, true, message)
}

func (c *Conn) writeServiceMessage(ctx context.Context, message bin.Encoder) error {
	return c.write(ctx, atomic.AddInt64(&c.reqID, 1), false, message)
}

func (c *Conn) write(ctx context.Context, reqID int64, content bool, message bin.Encoder) error {
	c.reqMux.Lock()
	defer c.reqMux.Unlock()

	// Note that reqID is internal RPC request id that can be only used for
	// tracing and has no relation to telegram server or protocol.
	msgID, ok := c.reqToMsg[reqID]
	if !ok {
		msgID = c.newMessageID()

		// Saving mapping of internal to external id.
		//
		// This will OOM eventually.
		// TODO(ernado): cleanup by callback from rpc engine
		c.reqToMsg[reqID] = msgID
		c.msgToReq[msgID] = reqID
	}

	cleanup := func() {
		delete(c.reqToMsg, reqID)
		delete(c.msgToReq, msgID)
	}

	// Computing current sequence number (seqno).
	// This should be serialized with new message id generation.
	//
	// See https://github.com/gotd/td/issues/245 for reference.
	seq := c.sentContentMessages * 2
	if content {
		seq++
		c.sentContentMessages++
	}

	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(msgID, seq, message, b); err != nil {
		cleanup()
		return err
	}
	if err := c.conn.Send(ctx, b); err != nil {
		cleanup()
		return err
	}

	return nil
}
