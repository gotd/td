package telegram

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/pool"
	"github.com/gotd/td/tg"
)

func (c *Client) createPool(dc int, max int64, creator func() pool.Conn) (*pool.DC, error) {
	select {
	case <-c.ctx.Done():
		return nil, xerrors.Errorf("client already closed: %w", c.ctx.Err())
	default:
	}

	p := pool.NewDC(c.ctx, dc, creator, pool.DCOptions{
		Logger:             c.log.Named("pool").With(zap.Int("dc_id", dc)),
		MaxOpenConnections: max,
	})

	return p, nil
}

// Pool creates new multi-connection invoker to current DC.
func (c *Client) Pool(max int64) (tg.Invoker, error) {
	if max < 0 {
		return nil, xerrors.Errorf("invalid max value %d", max)
	}

	s := c.session.Load()
	return c.createPool(s.DC, max, func() pool.Conn {
		return c.buildConn(connModeData, c.session).Build()
	})
}

func (c *Client) dc(ctx context.Context, id int, max int64) (*pool.DC, error) {
	if max < 0 {
		return nil, xerrors.Errorf("invalid max value %d", max)
	}

	cfg := c.cfg.Load()
	opts := c.opts

	dc, ok := findDC(cfg, id, true)
	if !ok {
		return nil, xerrors.Errorf("failed to find DC %d", id)
	}

	if dc.CDN {
		cdnCfg, err := c.tg.HelpGetCDNConfig(ctx)
		if err != nil {
			return nil, xerrors.Errorf("get CDN config: %w", err)
		}

		keys, err := parseCDNKeys(cdnCfg.PublicKeys...)
		if err != nil {
			return nil, xerrors.Errorf("parse CDN keys: %w", err)
		}

		opts.PublicKeys = keys
		// Zero key for CDN.
		opts.Key = crypto.AuthKey{}
		opts.Salt = 0
	}

	addr := fmt.Sprintf("%s:%d", dc.IPAddress, dc.Port)

	c.sessionsMux.Lock()
	session, ok := c.sessions[id]
	if !ok {
		session = pool.NewSyncSession(pool.Session{})
		c.sessions[id] = session
	}
	c.sessionsMux.Unlock()

	p, err := c.createPool(id, max, func() pool.Conn {
		return c.buildConn(connModeData, session).
			WithNoopHandler().
			WithOptions(id, addr, opts).
			Build()
	})
	if err != nil {
		return nil, xerrors.Errorf("create pool: %w", err)
	}

	if !dc.CDN {
		_, err = c.transfer(ctx, tg.NewClient(p), id)
		if err != nil {
			// Ignore case then we are not authorized.
			if unauthorized(err) {
				return p, nil
			}

			return nil, xerrors.Errorf("transfer: %w", err)
		}
	}

	return p, nil
}

// DC creates new multi-connection invoker to given DC.
func (c *Client) DC(ctx context.Context, id int, max int64) (tg.Invoker, error) {
	return c.dc(ctx, id, max)
}
