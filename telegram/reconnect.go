package telegram

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

func (c *Client) reconnect() error {
	c.log.Debug("Disconnected. Trying to re-connect")

	if err := c.connect(c.ctx); err != nil {
		return xerrors.Errorf("connect: %w", err)
	}

	go func() {
		if err := c.initConnection(c.ctx); err != nil {
			c.log.With(zap.Error(err)).Error("Failed to init connection after reconnect")
			return
		}

		c.log.Debug("Reconnected")

		if err := c.ensureState(c.ctx); err != nil {
			c.log.With(zap.Error(err)).Error("Failed to get state after reconnect")
			return
		}
	}()

	return nil
}
