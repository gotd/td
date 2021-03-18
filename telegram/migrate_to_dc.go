package telegram

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

func (c *Client) ensureRestart(ctx context.Context, export *tg.AuthExportedAuthorization) error {
	c.log.Debug("Triggering restart")
	c.resetReady()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.restart <- struct{}{}:
		c.log.Debug("Restart initialized")
	}

	if export != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c.exported <- export:
			c.log.Debug("Sent export authorization")
		}
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
		return c.invokeRaw(ctx, input, output)
	}

	mctx, cancel := context.WithTimeout(ctx, c.migrationTimeout)
	defer cancel()
	if err := c.migrateToDc(mctx, dcID,
		// TODO(tdakkota): Is it may be necessary to transfer auth
		//  if error is not FILE_MIGRATE or STATS_MIGRATE?
		false,
	); err != nil {
		return xerrors.Errorf("migrate to dc: %w", err)
	}

	// Re-trying request on another connection.
	return c.invokeRaw(ctx, input, output)
}

func (c *Client) migrateToDc(ctx context.Context, dcID int, transfer bool) error {
	cfg := c.cfg.Load()
	dcList := dcs.FindPrimaryDCs(cfg.DCOptions, dcID, false)
	if len(dcList) == 0 {
		return xerrors.Errorf("DC %d not found", dcID)
	}

	c.log.Info("Selected new DC from config",
		zap.Int("dc_id", dcID),
		zap.Int("candidates", len(dcList)),
	)

	var export *tg.AuthExportedAuthorization
	if transfer {
		exported, err := c.exportAuth(ctx, dcID)
		if err != nil && !unauthorized(err) {
			return xerrors.Errorf("export auth: %w", err)
		}
		export = exported
	}

	c.session.Migrate(dcID)
	return c.ensureRestart(ctx, export)
}
