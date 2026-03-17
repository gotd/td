package telegram

import (
	"context"
	"sync/atomic"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

// CloseInvoker is a closeable tg.Invoker.
type CloseInvoker interface {
	tg.Invoker
	Close() error
}

func (c *Client) createPool(dc int, max int64, creator func() pool.Conn) (*pool.DC, error) {
	select {
	case <-c.ctx.Done():
		return nil, errors.Wrap(c.ctx.Err(), "client already closed")
	default:
	}

	p := pool.NewDC(c.ctx, dc, creator, pool.DCOptions{
		Logger:             c.log.Named("pool").With(zap.Int("dc_id", dc)),
		MaxOpenConnections: max,
	})

	return p, nil
}

// Pool creates new multi-connection invoker to current DC.
func (c *Client) Pool(max int64) (CloseInvoker, error) {
	if max < 0 {
		return nil, errors.Errorf("invalid max value %d", max)
	}

	s := c.session.Load()
	return c.createPool(s.DC, max, func() pool.Conn {
		id := c.connsCounter.Inc()
		return c.createConn(id, manager.ConnModeData, nil, func(err error) {
			// Primary pool connections share persisted primary session state.
			c.handlePrimaryConnDead(err)
			if c.onDead != nil {
				c.onDead(err)
			}
		})
	})
}

func (c *Client) dc(
	ctx context.Context,
	dcID int,
	max int64,
	dialer mtproto.Dialer,
	mode manager.ConnMode,
) (*pool.DC, error) {
	if max < 0 {
		return nil, errors.Errorf("invalid max value %d", max)
	}

	dcList := dcs.FindDCs(c.cfg.Load().DCOptions, dcID, false)
	if len(dcList) < 1 {
		return nil, errors.Errorf("unknown DC %d", dcID)
	}
	c.log.Debug("Creating pool",
		zap.Int("dc_id", dcID),
		zap.Int64("max", max),
		zap.Int("candidates", len(dcList)),
	)

	opts := c.opts
	if mode == manager.ConnModeCDN {
		// TDesktop-compatible gate: CDN connection is allowed only when keyset
		// for requested CDN DC is present (or can be fetched).
		cdnKeys, set := c.cachedCDNKeysForDC(dcID)
		if !set || len(cdnKeys) == 0 {
			fetched, err := c.fetchCDNKeysForDC(ctx, dcID)
			if err != nil {
				return nil, errors.Wrapf(err, "fetch CDN public keys for DC %d", dcID)
			}
			cdnKeys = fetched
		}
		if len(cdnKeys) == 0 {
			return nil, errors.Errorf("no CDN public keys available for CDN DC %d", dcID)
		}
		// Keep CDN keys first and extend with bundled keys for fingerprint
		// compatibility fallback, matching TDesktop key lookup behavior.
		opts.PublicKeys = mergePublicKeys(cdnKeys, opts.PublicKeys)
	}
	// suppressSetup temporarily disables per-connection transfer hook while
	// explicit first transfer below is running, avoiding duplicate import.
	var suppressSetup atomic.Bool
	p, err := c.createPool(dcID, max, func() pool.Conn {
		id := c.connsCounter.Inc()

		c.sessionsMux.Lock()
		sessions := c.sessions
		if mode == manager.ConnModeCDN {
			// Keep CDN auth key lifecycle separated from regular DC sessions.
			sessions = c.cdnSessions
		}
		session, ok := sessions[dcID]
		if !ok {
			session = pool.NewSyncSession(pool.Session{DC: dcID})
			sessions[dcID] = session
		}
		c.sessionsMux.Unlock()

		options, data := session.Options(opts)
		setup := manager.SetupCallback(nil)
		handler := c.asHandler()
		if mode != manager.ConnModeCDN &&
			data.AuthKey.Zero() &&
			c.session.Load().DC != dcID &&
			!suppressSetup.Load() {
			// Non-main DC key must be authorized via auth.export/import after
			// local key generation.
			setup = c.dcTransferSetup(dcID)
		}
		if mode == manager.ConnModeCDN {
			// CDN pools do not process updates and use dedicated session store.
			handler = c.asCDNHandler()
		}
		options.Logger = c.log.Named("conn").With(
			zap.Int64("conn_id", id),
			zap.Int("dc_id", dcID),
		)
		return c.create(
			dialer, mode, c.appID,
			options, manager.ConnOptions{
				DC:      dcID,
				Device:  c.device,
				Handler: handler,
				Setup:   setup,
				OnDead: func(err error) {
					if mode == manager.ConnModeCDN {
						// CDN dead handler also manages CDN key invalidation.
						c.handleCDNConnDead(dcID, err)
						return
					}
					c.handleDCConnDead(dcID, err)
				},
			},
		)
	})
	if err != nil {
		return nil, errors.Wrap(err, "create pool")
	}

	if mode == manager.ConnModeCDN {
		// No auth transfer for CDN mode: CDN API uses file tokens and does not
		// require auth.export/import bootstrap.
		return p, nil
	}

	// First transfer is done explicitly to preserve old behavior: return
	// transfer errors from DC pool creation. Setup callback remains enabled for
	// future reconnections when keys are re-generated inside the pool.
	suppressSetup.Store(true)
	_, err = c.transfer(ctx, tg.NewClient(p), dcID)
	suppressSetup.Store(false)
	if err != nil {
		// Ignore case then we are not authorized.
		if auth.IsUnauthorized(err) {
			return p, nil
		}

		// Kill pool if we got error.
		_ = p.Close()
		return nil, errors.Wrap(err, "transfer")
	}

	return p, nil
}

// DC creates new multi-connection invoker to given DC.
func (c *Client) DC(ctx context.Context, dc int, max int64) (CloseInvoker, error) {
	return c.dc(ctx, dc, max, c.primaryDC(dc), manager.ConnModeData)
}

// MediaOnly creates new multi-connection invoker to given DC ID.
// It connects to MediaOnly DCs.
func (c *Client) MediaOnly(ctx context.Context, dc int, max int64) (CloseInvoker, error) {
	return c.dc(ctx, dc, max, func(ctx context.Context) (transport.Conn, error) {
		return c.resolver.MediaOnly(ctx, dc, c.dcList())
	}, manager.ConnModeData)
}

// CDN creates new multi-connection invoker to given CDN DC ID.
// It connects to CDN DCs.
func (c *Client) CDN(ctx context.Context, dc int, max int64) (CloseInvoker, error) {
	if max < 0 {
		return nil, errors.Errorf("invalid max value %d", max)
	}
	need := normalizeCDNPoolMax(max)

	if cached, ok := c.cdnPools.acquire(dc, need); ok {
		// Reuse existing pool to avoid extra TCP/MTProto handshakes.
		return cached, nil
	}

	// Keep shared CDN pools per DC with max-aware reuse.
	created, err := c.dc(ctx, dc, need, func(ctx context.Context) (transport.Conn, error) {
		return c.resolver.CDN(ctx, dc, c.dcList())
	}, manager.ConnModeCDN)
	if err != nil {
		return nil, err
	}

	handle, reused := c.cdnPools.publishOrAcquire(dc, need, created)
	if reused {
		// Lost race: another goroutine already published suitable pool.
		_ = created.Close()
		return handle, nil
	}
	return handle, nil
}
