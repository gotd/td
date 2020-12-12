package telegram

import (
	"context"

	"go.uber.org/zap"
)

func (c *Client) ensureState(ctx context.Context) error {
	c.log.Debug("Trying to get state")
	state, err := c.tg.UpdatesGetState(ctx)
	if err != nil {
		return err
	}
	c.log.With(
		zap.Int("pts", state.Pts),
		zap.Int("qts", state.Qts),
		zap.Int("seq", state.Seq),
		zap.Int("unread_count", state.UnreadCount),
	).Info("Got state")
	return nil
}
