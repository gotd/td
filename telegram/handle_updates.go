package telegram

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func (c *Client) processUpdates(updates tg.UpdatesClass) error {
	if c.updateHandler == nil {
		// Ignoring. Probably we should ACK.
		return nil
	}
	switch u := updates.(type) {
	case *tg.Updates:
		go func() {
			if c.updateHandler == nil {
				return
			}
			// We should send ACK here.
			if err := c.updateHandler(c.ctx, c, u); err != nil {
				c.log.With(zap.Error(err)).Error("Update handler returning error")
			}
		}()
		return nil
	default:
		c.log.With(zap.String("update_type", fmt.Sprintf("%T", u))).Debug("Ignoring update")
	}
	return nil
}

func (c *Client) handleUpdates(b *bin.Buffer) error {
	updates, err := tg.DecodeUpdates(b)
	if err != nil {
		return xerrors.Errorf("failed to decode updates: %w", err)
	}
	return c.processUpdates(updates)
}
