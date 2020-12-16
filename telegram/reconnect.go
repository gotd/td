package telegram

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

func (c *Client) reconnect() error {
	c.log.Debug("Disconnected. Trying to re-connect")

	// Probably we should set read or write deadline here.

	conn, err := c.dialer.DialContext(c.ctx, "tcp", c.addr)
	if err != nil {
		return xerrors.Errorf("dial: %w", err)
	}

	// TODO(ernado): data race possible for writes from other goroutines!
	c.conn = conn

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
