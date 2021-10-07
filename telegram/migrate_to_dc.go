package telegram

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
)

func (c *Client) ensureRestart(ctx context.Context) error {
	c.log.Debug("Triggering restart")
	c.resetReady()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.restart <- struct{}{}:
		c.log.Debug("Restart initialized")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ready.Ready():
		c.log.Info("Restart ensured")
		return nil
	}
}

func (c *Client) invokeMigrate(ctx context.Context, dcID int, input bin.Encoder, output bin.Decoder) error {
	// Acquire or cancel.
	select {
	case c.migration <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	// Release.
	defer func() {
		<-c.migration
	}()

	// Check if someone already migrated.
	s := c.session.Load()
	if s.DC == dcID {
		return c.invokeConn(ctx, input, output)
	}

	mctx, cancel := context.WithTimeout(ctx, c.migrationTimeout)
	defer cancel()
	if err := c.migrateToDc(mctx, dcID); err != nil {
		return xerrors.Errorf("migrate to dc: %w", err)
	}

	// Re-trying request on another connection.
	return c.invokeConn(ctx, input, output)
}

func (c *Client) migrateToDc(ctx context.Context, dcID int) error {
	c.session.Migrate(dcID)
	return c.ensureRestart(ctx)
}
