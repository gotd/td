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
	reqID := atomic.AddInt64(&c.reqID, 1)
	defer c.cleanup(reqID)

	return c.write(ctx, reqID, false, message)
}

func (c *Conn) write(ctx context.Context, reqID int64, content bool, message bin.Encoder) error {
	c.reqMux.Lock()

	// Note that reqID is internal RPC request id that can be only used for
	// tracing and has no relation to telegram server or protocol.
	msgID := c.reqToMsg[reqID]
	seq, ok := c.reqToSeq[reqID]
	if !ok {
		msgID = c.newMessageID()

		// Computing current sequence number (seqno).
		// This should be serialized with new message id generation.
		//
		// See https://github.com/gotd/td/issues/245 for reference.
		seq = c.sentContentMessages * 2
		if content {
			seq++
			c.sentContentMessages++
		}

		// Saving mapping of internal to external id.
		c.reqToMsg[reqID] = msgID
		c.msgToReq[msgID] = reqID

		// Note that sequence id should not change for that request on retry.
		c.reqToSeq[reqID] = seq
	}

	// It is not required to serialize on-the-wire writes (at least for now),
	// so we can release mutex here.
	c.reqMux.Unlock()

	b := new(bin.Buffer)
	if err := c.newEncryptedMessage(msgID, seq, message, b); err != nil {
		return err
	}
	if err := c.conn.Send(ctx, b); err != nil {
		return err
	}

	return nil
}

func (c *Conn) cleanup(reqID int64) {
	c.reqMux.Lock()
	defer c.reqMux.Unlock()

	msgID, ok := c.reqToMsg[reqID]
	delete(c.reqToMsg, reqID)
	delete(c.reqToSeq, reqID)
	if ok {
		delete(c.msgToReq, msgID)
	}
}
