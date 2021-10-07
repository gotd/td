package mtproto

import (
	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/proto"

	"github.com/nnqq/td/bin"
)

func (c *Conn) handleContainer(msgID int64, b *bin.Buffer) error {
	var container proto.MessageContainer
	if err := container.Decode(b); err != nil {
		return xerrors.Errorf("container: %w", err)
	}
	for _, msg := range container.Messages {
		if err := c.processContainerMessage(msgID, msg); err != nil {
			return err
		}
	}
	return nil
}

func (c *Conn) processContainerMessage(msgID int64, msg proto.Message) error {
	b := &bin.Buffer{Buf: msg.Body}
	return c.handleMessage(msgID, b)
}
