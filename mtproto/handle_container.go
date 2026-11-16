package mtproto

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/proto"
)

func (c *Conn) handleContainer(ctx context.Context, received time.Time, b *bin.Buffer) error {
	var container proto.MessageContainer
	if err := container.Decode(b); err != nil {
		return errors.Wrap(err, "container")
	}
	for _, msg := range container.Messages {
		if err := c.processContainerMessage(ctx, received, msg); err != nil {
			return err
		}
	}
	return nil
}

// processContainerMessage handles a single message of a received container.
//
// The container itself is not content-related (even seqno), so content-related
// inner messages (odd seqno) must be acknowledged individually, like tdesktop
// and tdlib do.
//
// Inner msg_ids are checked for duplicates like top-level ones (tdesktop
// registers them in its received ids, Android checks isMessageIdProcessed): a
// message the server sends again, e.g. in a new container after our ack was
// lost, is acknowledged again but not handled twice.
func (c *Conn) processContainerMessage(ctx context.Context, received time.Time, msg proto.Message) error {
	var handleErr error
	if c.messageIDBuf == nil || c.messageIDBuf.Consume(msg.ID) {
		handleErr = c.handleMessage(ctx, received, msg.ID, &bin.Buffer{Buf: msg.Body})
	} else {
		c.traceRejected(msg.ID, "bad_msg_id")
		c.log.Debug(ctx, "Ignoring duplicate message in container", log.Int64("msg_id", msg.ID))
	}
	// Acknowledging even on handling error or duplicate, same as for
	// top-level messages.
	if needAck(msg.SeqNo) {
		if err := c.enqueueAck(ctx, msg.ID); err != nil {
			return err
		}
	}
	return handleErr
}
