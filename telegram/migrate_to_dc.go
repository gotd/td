package telegram

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"golang.org/x/xerrors"

	"github.com/gotd/td/mtproto"
)

func (c *Client) migrateToDc(ctx context.Context, dcID int) error {
	c.log.Info("Migrating to another DC", zap.Int("dc", dcID))
	c.connMux.Lock()
	defer c.connMux.Unlock()

	cfg := c.conn.Config()

	for _, dc := range cfg.DCOptions {
		if dc.ID != dcID {
			continue
		}
		if dc.MediaOnly || dc.Ipv6 || dc.CDN || dc.TcpoOnly {
			continue
		}

		c.connOpt.Addr = fmt.Sprintf("%s:%d", dc.IPAddress, dc.Port)
		c.log.Info("Selected new addr from config",
			zap.String("addr", c.connOpt.Addr),
		)
	}

	// Swapping connections.
	if err := c.conn.Close(); err != nil {
		c.log.Warn("Failed to close old connection", zap.Error(err))
	}
	c.conn = mtproto.NewConn(c.appID, c.appHash, c.connOpt)
	if err := c.conn.Connect(ctx); err != nil {
		return xerrors.Errorf("connect: %w", err)
	}

	return nil
}
