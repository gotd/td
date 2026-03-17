package telegram

import (
	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
)

func (c *Client) handlePrimaryConnDead(err error) {
	if !errors.Is(err, mtproto.ErrPFSDropKeysRequired) {
		return
	}

	// Keep DC but wipe key/salt so next reconnect performs full key bootstrap.
	dc := c.session.Load().DC
	c.log.Warn("Dropping stored primary session key after PFS key reset request",
		zap.Int("dc_id", dc),
	)
	c.session.Store(pool.Session{DC: dc})

	c.sessionsMux.Lock()
	if s, ok := c.sessions[dc]; ok {
		s.Store(pool.Session{DC: dc})
	}
	c.sessionsMux.Unlock()
}

func (c *Client) handleDCConnDead(dcID int, err error) {
	if !errors.Is(err, mtproto.ErrPFSDropKeysRequired) {
		// Preserve old error path for non-PFS failures.
		if c.onDead != nil {
			c.onDead(err)
		}
		return
	}

	c.log.Warn("Dropping stored DC session key after PFS key reset request",
		zap.Int("dc_id", dcID),
	)
	c.sessionsMux.Lock()
	s, ok := c.sessions[dcID]
	if !ok {
		s = pool.NewSyncSession(pool.Session{DC: dcID})
		c.sessions[dcID] = s
	}
	s.Store(pool.Session{DC: dcID})
	c.sessionsMux.Unlock()

	if c.onDead != nil {
		c.onDead(err)
	}
}
