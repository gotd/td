package telegram

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tg"
)

func (c *Client) restoreConnection(ctx context.Context) error {
	if c.storage == nil {
		return nil
	}

	data, err := c.storage.Load(ctx)
	if errors.Is(err, session.ErrNotFound) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "load")
	}

	// If file does not contain DC ID, so we use DC from options.
	prev := c.session.Load()
	if data.DC == 0 {
		data.DC = prev.DC
	}

	// Restoring persisted auth key.
	var key crypto.AuthKey
	copy(key.Value[:], data.AuthKey)
	copy(key.ID[:], data.AuthKeyID)

	if key.Value.ID() != key.ID {
		return errors.New("corrupted key")
	}

	// Re-initializing connection from persisted state.
	c.log.Info("Connection restored from state",
		zap.String("addr", data.Addr),
		zap.String("key_id", fmt.Sprintf("%x", data.AuthKeyID)),
	)

	c.connMux.Lock()
	c.session.Store(pool.Session{
		DC:      data.DC,
		AuthKey: key,
		Salt:    data.Salt,
	})
	c.conn = c.createPrimaryConn(nil)
	c.connMux.Unlock()

	return nil
}

func (c *Client) saveSession(cfg tg.Config, s mtproto.Session) error {
	if c.storage == nil {
		return nil
	}

	data, err := c.storage.Load(c.ctx)
	if errors.Is(err, session.ErrNotFound) {
		// Initializing new state.
		err = nil
		data = &session.Data{}
	}
	if err != nil {
		return errors.Wrap(err, "load")
	}

	// Updating previous data.
	data.Config = session.ConfigFromTG(cfg)
	keyToSave := s.Key
	if !s.PermKey.Zero() {
		// Persist permanent key in PFS mode: temporary key is expected to rotate.
		keyToSave = s.PermKey
	}
	data.AuthKey = keyToSave.Value[:]
	data.AuthKeyID = keyToSave.ID[:]
	data.DC = cfg.ThisDC
	data.Salt = s.Salt

	if err := c.storage.Save(c.ctx, data); err != nil {
		return errors.Wrap(err, "save")
	}

	c.log.Debug("Data saved",
		zap.String("key_id", fmt.Sprintf("%x", data.AuthKeyID)),
	)
	return nil
}

func (c *Client) onSession(cfg tg.Config, s mtproto.Session) error {
	sessionData := dcSessionFromMTProto(cfg.ThisDC, s)
	// Track per-DC session in memory for pool reconnections/migrations.
	c.storeDCSess(c.sessions, sessionData)

	primaryDC := c.session.Load().DC
	// Do not save session for non-primary DC.
	if cfg.ThisDC != 0 && primaryDC != 0 && primaryDC != cfg.ThisDC {
		return nil
	}

	c.connMux.Lock()
	c.session.Store(sessionData)
	c.cfg.Store(cfg)
	c.onReady()
	c.connMux.Unlock()

	if err := c.saveSession(cfg, s); err != nil {
		return errors.Wrap(err, "save")
	}

	return nil
}

func (c *Client) onCDNSession(cfg tg.Config, s mtproto.Session) error {
	// CDN sessions are isolated from regular DC map because lifecycle and reset
	// triggers differ (fingerprint misses, CDN-specific reconnects).
	c.storeDCSess(c.cdnSessions, dcSessionFromMTProto(cfg.ThisDC, s))
	return nil
}

func (c *Client) storeDCSess(target map[int]*pool.SyncSession, data pool.Session) {
	c.sessionsMux.Lock()
	if existing, ok := target[data.DC]; ok {
		existing.Store(data)
		c.sessionsMux.Unlock()
		return
	}
	target[data.DC] = pool.NewSyncSession(data)
	c.sessionsMux.Unlock()
}

func dcSessionFromMTProto(dc int, s mtproto.Session) pool.Session {
	keyToStore := s.Key
	if !s.PermKey.Zero() {
		// Keep in-memory/persisted key format backward-compatible: one key slot.
		// In PFS mode temp key rotates, so we pin permanent key.
		keyToStore = s.PermKey
	}

	return pool.Session{
		DC:      dc,
		Salt:    s.Salt,
		AuthKey: keyToStore,
	}
}
