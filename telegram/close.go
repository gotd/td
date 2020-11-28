package telegram

import (
	"context"

	"golang.org/x/xerrors"
)

// Close closes underlying connection and saves session to storage
// if provided.
func (c *Client) Close(ctx context.Context) error {
	c.cancel()

	if err := c.conn.Close(); err != nil {
		return err
	}

	// Probably we should wait with timeout, but it is unclear
	// whether we can try to save session or should hard fail on timeout.
	c.wg.Wait()

	if err := c.saveSession(ctx); err != nil {
		return xerrors.Errorf("failed to save session: %w", err)
	}

	return nil
}
