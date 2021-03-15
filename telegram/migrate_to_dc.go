package telegram

import (
	"context"
	"net"
	"sort"
	"strconv"

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

func findDC(cfg tg.Config, dcID int, preferIPv6 bool) (tg.DCOption, bool) {
	// Preallocate slice.
	candidates := make([]int, 0, 32)

	opts := cfg.DCOptions
	for idx, candidateDC := range opts {
		if candidateDC.ID != dcID {
			continue
		}
		candidates = append(candidates, idx)
	}

	if len(candidates) < 1 {
		return tg.DCOption{}, false
	}

	sort.Slice(candidates, func(i, j int) bool {
		l, r := opts[candidates[i]], opts[candidates[j]]

		// If we prefer IPv6 and left is IPv6 and right is not, so then
		// left is smaller (would be before right).
		if preferIPv6 {
			if l.Ipv6 && !r.Ipv6 {
				return true
			}
			if !l.Ipv6 && r.Ipv6 {
				return false
			}
		}

		// Also we prefer static addresses.
		return l.Static && !r.Static
	})

	return opts[candidates[0]], true
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
	dc, ok := findDC(c.cfg.Load(), dcID, c.opts.PreferIPv6)
	if !ok {
		return xerrors.Errorf("failed to find DC %d", dcID)
	}

	if dc.TCPObfuscatedOnly {
		return xerrors.Errorf("can't migrate to obfuscated transport only DC %d", dcID)
	}

	if dc.MediaOnly || dc.CDN {
		return xerrors.Errorf("can't migrate to CDN/Media-only DC %d", dcID)
	}

	addr := net.JoinHostPort(dc.IPAddress, strconv.Itoa(dc.Port))
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
	return c.ensureRestart(ctx, export)
}
