package telegram

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

func (c *Client) reconnect() error {
	// stop all network activity
	c.wcancel()
	c.wg.Wait()

	c.log.Debug("Disconnected. Trying to re-connect")

	// Probably we should set read or write deadline here.

	conn, err := c.dialer.DialContext(c.ctx, "tcp", c.addr)
	if err != nil {
		return xerrors.Errorf("failed to dial: %w", err)
	}

	// TODO(ernado): data race possible for writes from other goroutines!
	c.conn = conn

	if err := c.connect(c.ctx); err != nil {
		return xerrors.Errorf("failed to connect: %w", err)
	}

	c.wctx, c.wcancel = context.WithCancel(c.ctx)
	// init goroutines
	go c.readLoop(c.wctx)
	go c.writeLoop(c.wctx)

	if err := c.initConnection(c.ctx); err != nil {
		c.log.With(zap.Error(err)).Error("Failed to init connection after reconnect")
		return err
	}

	c.log.Debug("Reconnected")

	if err := c.ensureState(c.ctx); err != nil {
		c.log.With(zap.Error(err)).Error("Failed to get state after reconnect")
		return err
	}

	return nil
}
