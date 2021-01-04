package telegram

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

func (c *Client) createConn(addr string, mode connMode) clientConn {
	return newConn(addr, c.appID, c.appHash, mode, c.connOpt)
}

func (c *Client) migrateToDc(ctx context.Context, dcID int) error {
	c.log.Info("Migrating to another DC", zap.Int("dc", dcID))
	c.connMux.Lock()
	defer c.connMux.Unlock()

	cfg := c.conn.Config()

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

	// Swapping connections.
	if err := c.conn.Close(); err != nil {
		c.log.Warn("Failed to close old connection", zap.Error(err))
	}
	c.conn = c.createConn(addr, connModeUpdates)
	if err := c.conn.Connect(ctx); err != nil {
		return xerrors.Errorf("connect: %w", err)
	}

	return nil
}
