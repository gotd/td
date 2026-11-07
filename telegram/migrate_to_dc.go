package telegram

import (
	"context"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
)

// ensureRestart restarts the primary connection and waits until it is ready.
// Reason is restartTimeout or restartMigration, for metrics and logs.
func (c *Client) ensureRestart(ctx context.Context, reason string) error {
	c.log.Debug(ctx, "Triggering restart", log.String("reason", reason))
	c.resetReady()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.restart <- struct{}{}:
		// Counted once sent, not when abandoned on cancel.
		c.meters().restarts.Add(ctx, 1, metric.WithAttributes(attribute.String("reason", reason)))
		c.log.Debug(ctx, "Restart initialized")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ready.Ready():
		c.log.Info(ctx, "Restart ensured")
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
		return errors.Wrap(err, "migrate to dc")
	}

	// Re-trying request on another connection.
	return c.invokeConn(ctx, input, output)
}

func (c *Client) migrateToDc(ctx context.Context, dcID int) error {
	from := c.session.Load().DC
	c.log.Info(ctx, "Migrating primary connection",
		log.Int("from_dc", from),
		log.Int("dc_id", dcID),
	)
	c.meters().migrations.Add(ctx, 1, dcAttrs(dcID))
	c.session.Migrate(dcID)
	return c.ensureRestart(ctx, restartMigration)
}

// MigrateTo forces client to migrate to another DC.
func (c *Client) MigrateTo(ctx context.Context, dcID int) error {
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

	return c.migrateToDc(ctx, dcID)
}
