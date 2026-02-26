package mtproto

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
)

var (
	// ErrPFSReconnectRequired is returned when connection should be recreated
	// to continue PFS flow (e.g. temp key renewal).
	ErrPFSReconnectRequired = errors.New("pfs reconnect required")
	// ErrPFSDropKeysRequired is returned when stored permanent key should be
	// dropped and generated again on the next connection attempt.
	ErrPFSDropKeysRequired = errors.New("pfs drop keys required")
)

func (c *Conn) tempKeyRenewalLoop(ctx context.Context) error {
	if !c.pfs {
		<-ctx.Done()
		return ctx.Err()
	}

	for {
		c.sessionMux.RLock()
		expiresAt := c.tempKeyExpiry
		c.sessionMux.RUnlock()
		if expiresAt == 0 {
			// If expiry was not recorded (e.g. restored runtime state), derive
			// a conservative local horizon from configured ttl.
			expiresAt = c.clock.Now().Unix() + int64(c.tempKeyTTL)
		}

		// Reconnect after 75% of ttl to leave buffer for bind retries.
		renewAt := expiresAt - int64(c.tempKeyTTL)/4
		wait := time.Unix(renewAt, 0).Sub(c.clock.Now())
		if wait < 0 {
			wait = 0
		}

		timer := c.clock.Timer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C():
		}
		c.log.Info("Temporary auth key renewal required, reconnecting",
			zap.Int64("expires_at", expiresAt),
			zap.Int64("renew_at", renewAt),
		)
		return errors.Wrap(ErrPFSReconnectRequired, "temporary auth key renewal required")
	}
}
