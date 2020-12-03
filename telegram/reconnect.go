package telegram

import "golang.org/x/xerrors"

func (c *Client) reconnect() error {
	c.log.Debug("Disconnected. Trying to re-connect.")

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

	if err := c.initConnection(c.ctx); err != nil {
		return xerrors.Errorf("failed to init connection: %w", err)
	}

	return nil
}
