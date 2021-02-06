package telegram

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
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

func findDC(cfg tg.Config, dcID int) (dc tg.DcOption, ok bool) {
	for _, dc := range cfg.DCOptions {
		if dc.ID == dcID {
			return dc, true
		}
	}

	ok = false
	return
}

func (c *Client) migrateToDc(ctx context.Context, dcID int) error {
	dc, ok := findDC(c.cfg.Load(), dcID)
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

	c.session.Migrate(dcID, addr)
	return c.ensureRestart(ctx)
}
