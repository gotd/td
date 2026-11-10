package mtproto

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/proto"
)

func (c *Conn) handleContainer(ctx context.Context, b *bin.Buffer) error {
	var container proto.MessageContainer
	if err := container.Decode(b); err != nil {
		return errors.Wrap(err, "container")
	}
	for _, msg := range container.Messages {
		if err := c.processContainerMessage(ctx, msg); err != nil {
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
// Inner msg_ids are not deduplicated via messageIDBuf yet: the buffer
// effectively keeps a single id (parity ID-4), so it would not be effective.
func (c *Conn) processContainerMessage(ctx context.Context, msg proto.Message) error {
	handleErr := c.handleMessage(ctx, msg.ID, &bin.Buffer{Buf: msg.Body})
	// Acknowledging even on handling error, same as for top-level messages.
	if needAck(msg.SeqNo) {
		if err := c.enqueueAck(ctx, msg.ID); err != nil {
			return err
		}
	}
	return handleErr
}
