package telegram

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
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

func findDC(cfg tg.Config, dcID int, noIPv6 bool) (dc tg.DcOption, ok bool) {
	for _, dc := range cfg.DCOptions {
		if noIPv6 && dc.Ipv6 {
			continue
		}

		if dc.ID == dcID {
			return dc, true
		}
	}

	ok = false
	return
}

func (c *Client) invokeMigrate(ctx context.Context, dcID int, input bin.Encoder, output bin.Decoder) error {
	c.migration.Lock()
	defer c.migration.Unlock()

	// Check if someone already migrated.
	s := c.session.Load()
	if s.DC == dcID {
		return c.invokeRaw(ctx, input, output)
	}

	if err := c.migrateToDc(
		c.ctx, dcID,
		// TODO(tdakkota): Is it may be necessary to migrate if error is not FILE_MIGRATE or STATS_MIGRATE?
		false,
	); err != nil {
		return xerrors.Errorf("migrate to dc: %w", err)
	}

	// Re-trying request on another connection.
	return c.invokeRaw(ctx, input, output)
}

func (c *Client) migrateToDc(ctx context.Context, dcID int, transfer bool) error {
	dc, ok := findDC(c.cfg.Load(), dcID, true)
	if !ok {
		return xerrors.Errorf("failed to find DC %d", dcID)
	}

	if dc.TcpoOnly {
		return xerrors.Errorf("can't migrate to obfuscated transport only DC %d", dcID)
	}

	if dc.MediaOnly || dc.CDN {
		return xerrors.Errorf("can't migrate to CDN/Media-only DC %d", dcID)
	}

	addr := fmt.Sprintf("%s:%d", dc.IPAddress, dc.Port)
	c.log.Info("Selected new addr from config", zap.String("addr", addr))

	var export *tg.AuthExportedAuthorization
	if transfer {
		var err error
		export, err = c.exportAuth(ctx, dcID)
		if err != nil {
			if !unauthorized(err) {
				c.log.Info("Export authorization failed", zap.Error(err))
			}
		}
	}

	c.session.Migrate(dcID, addr)
	c.primaryDC.Store(int64(dcID))
	return c.ensureRestart(ctx, export)
}
