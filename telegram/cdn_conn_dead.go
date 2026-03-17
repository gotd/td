package telegram

import (
	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/exchange"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
)

func (c *Client) handleCDNConnDead(dcID int, err error) {
	if errors.Is(err, exchange.ErrKeyFingerprintNotFound) {
		c.log.Warn("Resetting cached CDN keys after fingerprint miss",
			zap.Int("dc_id", dcID),
		)
		c.cdnKeysMux.Lock()
		c.cdnKeys = nil
		c.cdnKeysByDC = nil
		c.cdnKeysSet = false
		// Bump generation so in-flight help.getCdnConfig results are discarded if
		// they were started before invalidation.
		c.cdnKeysGen++
		c.cdnKeysMux.Unlock()
		// Drop current singleflight entry so next attempt refetches keys.
		c.cdnKeysLoad.Forget(helpGetCDNConfigSingleflightKey)

		// Close asynchronously: callback may be invoked from pool worker
		// goroutine, and synchronous self-close can deadlock on Wait().
		// Queue closes through bounded workers to avoid unbounded goroutine fan-out.
		c.cdnPools.invalidateDC(dcID)
		// Fingerprint miss is recoverable and handled internally by invalidation
		// + reconnect with fresh keys, no external onDead signal is needed.
		return
	}

	if !errors.Is(err, mtproto.ErrPFSDropKeysRequired) {
		// Keep legacy callback semantics for all non-PFS errors.
		if c.onDead != nil {
			c.onDead(err)
		}
		return
	}

	c.log.Warn("Dropping stored CDN session key after PFS key reset request",
		zap.Int("dc_id", dcID),
	)
	c.sessionsMux.Lock()
	s, ok := c.cdnSessions[dcID]
	if !ok {
		s = pool.NewSyncSession(pool.Session{DC: dcID})
		c.cdnSessions[dcID] = s
	}
	s.Store(pool.Session{DC: dcID})
	c.sessionsMux.Unlock()

	if c.onDead != nil {
		c.onDead(err)
	}
}
