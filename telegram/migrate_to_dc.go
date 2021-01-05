package telegram

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
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
	case <-c.ready:
		c.log.Info("Restart ensured")
		return nil
	}
}

func (c *Client) migrateToDc(ctx context.Context, dcID int) error {
	c.connMux.Lock()
	cfg := c.cfg
	c.connMux.Unlock()

	var addr string
	for _, dc := range cfg.DCOptions {
		if dc.ID != dcID {
			continue
		}
		if dc.MediaOnly || dc.Ipv6 || dc.CDN || dc.TcpoOnly {
			continue
		}
		addr = fmt.Sprintf("%s:%d", dc.IPAddress, dc.Port)
		c.log.Info("Selected new addr from config", zap.String("addr", addr))
		break
	}

	if addr == "" {
		return xerrors.Errorf("failed to find addr for dc %d", dcID)
	}

	c.connMux.Lock()
	c.connAddr = addr
	c.connMux.Unlock()

	return c.ensureRestart(ctx)
}
