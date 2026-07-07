package mtproto

import (
	"context"
	"time"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
)

func (c *Conn) writeContentMessage(ctx context.Context, msgID int64, seqNo int32, message bin.Encoder) error {
	return c.write(ctx, msgID, seqNo, message)
}

func (c *Conn) writeServiceMessage(ctx context.Context, message bin.Encoder) error {
	msgID, seqNo := c.nextMsgSeq(false)
	return c.write(ctx, msgID, seqNo, message)
}

var bufPool = bin.NewPool(0)

// slowWriteThreshold is the duration after which a single write (lock
// acquisition + encrypt + transport send) is considered suspiciously slow and
// logged at warning level. A blocked or slow write is the primary signature of
// a half-open connection where updates are still received but requests can no
// longer be issued.
const slowWriteThreshold = 3 * time.Second

func (c *Conn) write(ctx context.Context, msgID int64, seqNo int32, message bin.Encoder) error {
	start := c.clock.Now()

	// Grab shared lock for writing.
	// It prevents message sending during key regeneration if server forgot current auth key.
	//
	// Note: if key exchange holds the write lock (see createAuthKey), every
	// send — including pings, acks and content requests — blocks here. Tracing
	// the lock acquisition separately lets us distinguish "stuck on exchange
	// lock" from "stuck on transport send".
	c.exchangeLock.RLock()
	defer c.exchangeLock.RUnlock()

	locked := c.clock.Now()
	if waited := locked.Sub(start); waited > slowWriteThreshold {
		c.logWithTypeID(peekID(message)).Warn(ctx, "Slow write: waited for exchange lock",
			log.Int64("msg_id", msgID),
			log.Int32("seq_no", seqNo),
			log.Duration("waited", waited),
		)
	}

	b := bufPool.Get()
	defer bufPool.Put(b)

	if err := c.newEncryptedMessage(msgID, seqNo, message, b); err != nil {
		return err
	}

	logger := c.logWithTypeID(peekID(message)).With(
		log.Int64("msg_id", msgID),
		log.Int32("seq_no", seqNo),
		log.Int("size_bytes", b.Len()),
	)
	logger.Debug(ctx, "Sending message")

	if err := c.conn.Send(ctx, b); err != nil {
		logger.Debug(ctx, "Send failed", log.Error(err), log.Duration("elapsed", c.clock.Now().Sub(locked)))
		return err
	}

	if elapsed := c.clock.Now().Sub(locked); elapsed > slowWriteThreshold {
		logger.Warn(ctx, "Slow write: transport send took too long", log.Duration("elapsed", elapsed))
	}

	return nil
}

// peekID returns the type id of an encodable message for logging, or 0 if it
// cannot be determined without encoding.
func peekID(message bin.Encoder) uint32 {
	if t, ok := message.(interface{ TypeID() uint32 }); ok {
		return t.TypeID()
	}
	return 0
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
