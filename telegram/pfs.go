package telegram

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/log"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
)

func (c *Client) handlePrimaryConnDead(err error) {
	if !errors.Is(err, mtproto.ErrPFSDropKeysRequired) {
		return
	}

	// Keep DC but wipe key/salt so next reconnect performs full key bootstrap.
	dc := c.session.Load().DC
	c.log.Warn(context.Background(), "Dropping stored primary session key after PFS key reset request",
		log.Int("dc_id", dc),
	)
	c.session.Store(pool.Session{DC: dc})

	c.sessionsMux.Lock()
	if s, ok := c.sessions[dc]; ok {
		c.dropKeyLocked(dc, s)
	}
	c.sessionsMux.Unlock()
}

// dropKeyLocked wipes the key and salt of the stored session of dc and
// remembers the dropped key, so that connections still running on it don't
// store it back. Must be called with sessionsMux held.
func (c *Client) dropKeyLocked(dc int, s *pool.SyncSession) {
	if old := s.Load(); !old.AuthKey.Zero() {
		c.droppedKeys[dc] = old.AuthKey.ID
	}
	delete(c.dcAuthorized, dc)
	s.Store(pool.Session{DC: dc})
}

func (c *Client) handleDCConnDead(dcID int, err error) {
	if !errors.Is(err, mtproto.ErrPFSDropKeysRequired) {
		// Preserve old error path for non-PFS failures.
		c.notifyConnDead(ConnDeadEvent{DC: dcID, Kind: ConnKindData, Err: err})
		return
	}

	c.log.Warn(context.Background(), "Dropping stored DC session key after PFS key reset request",
		log.Int("dc_id", dcID),
	)
	c.sessionsMux.Lock()
	s, ok := c.sessions[dcID]
	if !ok {
		s = pool.NewSyncSession(pool.Session{DC: dcID})
		c.sessions[dcID] = s
	}
	c.dropKeyLocked(dcID, s)
	c.sessionsMux.Unlock()

	c.notifyConnDead(ConnDeadEvent{DC: dcID, Kind: ConnKindData, Err: err})
}
