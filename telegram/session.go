package telegram

import (
	"bytes"
	"context"
	"fmt"

	"github.com/go-faster/errors"

	"github.com/gotd/log"

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
	c.log.Info(ctx, "Connection restored from state",
		log.String("addr", data.Addr),
		log.String("key_id", fmt.Sprintf("%x", data.AuthKeyID)),
	)

	c.connMux.Lock()
	c.session.Store(pool.Session{
		DC:           data.DC,
		AuthKey:      key,
		Salt:         data.Salt,
		KeyCreatedAt: data.AuthKeyCreatedAt,
	})
	c.replaceConn(c.createPrimaryConn(nil))
	c.connMux.Unlock()

	c.restoreDCKeys(ctx, data)
	if data.UserID != 0 {
		c.selfID.Store(data.UserID)
		c.selfBot.Store(data.Bot)
	}

	return nil
}

// restoreDCKeys loads authorized keys of the other DCs, so that their pools
// skip key exchange and authorization transfer.
func (c *Client) restoreDCKeys(ctx context.Context, data *session.Data) {
	for dc, k := range data.DCs {
		if dc == data.DC || !k.Authorized {
			continue
		}
		var key crypto.AuthKey
		copy(key.Value[:], k.AuthKey)
		copy(key.ID[:], k.AuthKeyID)
		if key.Zero() || key.Value.ID() != key.ID {
			c.log.Warn(ctx, "Skipping corrupted stored DC key", log.Int("dc_id", dc))
			continue
		}
		c.sessionsMux.Lock()
		c.sessions[dc] = pool.NewSyncSession(pool.Session{
			DC:           dc,
			AuthKey:      key,
			Salt:         k.Salt,
			KeyCreatedAt: k.CreatedAt,
		})
		c.dcAuthorized[dc] = key.ID
		c.sessionsMux.Unlock()
	}
}

// fillDCKeys sets data.DCs to the authorized keys of the other DCs.
func (c *Client) fillDCKeys(data *session.Data) {
	c.sessionsMux.Lock()
	defer c.sessionsMux.Unlock()
	keys := map[int]session.DCKey{}
	for dc, id := range c.dcAuthorized {
		s, ok := c.sessions[dc]
		if !ok || dc == data.DC {
			continue
		}
		v := s.Load()
		if v.AuthKey.Zero() || v.AuthKey.ID != id {
			continue
		}
		keys[dc] = session.DCKey{
			AuthKey:    append([]byte(nil), v.AuthKey.Value[:]...),
			AuthKeyID:  append([]byte(nil), v.AuthKey.ID[:]...),
			CreatedAt:  v.KeyCreatedAt,
			Authorized: true,
			Salt:       v.Salt,
		}
	}
	if len(keys) == 0 {
		keys = nil
	}
	data.DCs = keys
}

// updateStoredSession loads the stored session, applies f and saves it. It
// does nothing if there is no storage, no stored session yet, or after
// LogOut.
func (c *Client) updateStoredSession(ctx context.Context, f func(data *session.Data)) error {
	if c.storage == nil || c.auth.loggedOut.Load() {
		return nil
	}
	c.saveMux.Lock()
	defer c.saveMux.Unlock()
	data, err := c.storage.Load(ctx)
	if errors.Is(err, session.ErrNotFound) {
		// The primary session is saved first; it will include this.
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "load")
	}
	f(data)
	if err := c.storage.Save(ctx, data); err != nil {
		return errors.Wrap(err, "save")
	}
	return nil
}

// dcKeyAuthorized reports whether the stored key of dc already has the
// authorization transferred.
func (c *Client) dcKeyAuthorized(dc int) bool {
	c.sessionsMux.Lock()
	defer c.sessionsMux.Unlock()
	s, ok := c.sessions[dc]
	if !ok {
		return false
	}
	id, ok := c.dcAuthorized[dc]
	key := s.Load().AuthKey
	return ok && !key.Zero() && key.ID == id
}

// setDCAuthorized marks the current key of dc as authorized (or not) and
// persists the authorized keys.
func (c *Client) setDCAuthorized(ctx context.Context, dc int, authorized bool) {
	c.sessionsMux.Lock()
	changed := false
	if authorized {
		if s, ok := c.sessions[dc]; ok {
			if key := s.Load().AuthKey; !key.Zero() && c.dcAuthorized[dc] != key.ID {
				c.dcAuthorized[dc] = key.ID
				changed = true
			}
		}
	} else if _, ok := c.dcAuthorized[dc]; ok {
		delete(c.dcAuthorized, dc)
		changed = true
	}
	c.sessionsMux.Unlock()
	if !changed {
		return
	}
	if err := c.updateStoredSession(ctx, c.fillDCKeys); err != nil {
		c.log.Warn(ctx, "Failed to save DC keys", log.Int("dc_id", dc), log.Error(err))
	}
}

// acquireSessionState returns a free server session slot of dc for a new
// connection, see mtproto.SessionState.
func (c *Client) acquireSessionState(dc int) *mtproto.SessionState {
	c.sessionsMux.Lock()
	defer c.sessionsMux.Unlock()
	for _, s := range c.sessionStates[dc] {
		if s.TryAcquire() {
			return s
		}
	}
	s := mtproto.NewSessionState()
	s.TryAcquire()
	c.sessionStates[dc] = append(c.sessionStates[dc], s)
	return s
}

// onPrimaryNewSession recovers updates after the server started a new
// session on the primary connection: updates sent to the previous session
// may be lost, so the update handler gets updatesTooLong (getDifference in
// telegram/updates), like TDLib and Telegram Desktop. The first session of
// the client is skipped: there is no previous one.
func (c *Client) onPrimaryNewSession() {
	if c.primarySessions.Inc() == 1 || c.noUpdatesMode || c.updateHandler == nil {
		return
	}
	ctx := c.ctx
	if ctx == nil {
		return
	}
	c.log.Info(ctx, "New server session on primary connection, recovering updates")
	go func() {
		if err := c.updateHandler.Handle(ctx, &tg.UpdatesTooLong{}); err != nil {
			c.log.Debug(ctx, "Update handler failed on updatesTooLong", log.Error(err))
		}
	}()
}

// SelfID returns the ID of the authorized user or bot and whether it is a
// bot, without a request: ok is false if it is not known yet. It is known
// after a successful Self call (including the one Run makes on connect) and
// is restored from the stored session.
func (c *Client) SelfID() (id int64, bot bool, ok bool) {
	id = c.selfID.Load()
	return id, c.selfBot.Load(), id != 0
}

// setSelf records the authorized user and persists it.
func (c *Client) setSelf(ctx context.Context, u *tg.User) {
	if u == nil || u.ID == 0 {
		return
	}
	if c.selfID.Load() == u.ID && c.selfBot.Load() == u.Bot {
		return
	}
	c.selfID.Store(u.ID)
	c.selfBot.Store(u.Bot)
	if err := c.updateStoredSession(ctx, func(data *session.Data) {
		data.UserID = u.ID
		data.Bot = u.Bot
	}); err != nil {
		c.log.Warn(ctx, "Failed to save self ID", log.Error(err))
	}
}

func (c *Client) saveSession(cfg tg.Config, s mtproto.Session) error {
	if c.storage == nil || c.auth.loggedOut.Load() {
		// After LogOut the stored session stays wiped until the client is
		// authorized again.
		return nil
	}

	c.saveMux.Lock()
	defer c.saveMux.Unlock()
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
	switch {
	case s.KeyCreatedAt != 0:
		data.AuthKeyCreatedAt = s.KeyCreatedAt
	case !bytes.Equal(data.AuthKeyID, keyToSave.ID[:]):
		// A different key of unknown age.
		data.AuthKeyCreatedAt = 0
	}
	data.AuthKey = keyToSave.Value[:]
	data.AuthKeyID = keyToSave.ID[:]
	data.DC = cfg.ThisDC
	data.Salt = s.Salt
	c.fillDCKeys(data)
	if id := c.selfID.Load(); id != 0 {
		data.UserID = id
		data.Bot = c.selfBot.Load()
	}

	if err := c.storage.Save(c.ctx, data); err != nil {
		return errors.Wrap(err, "save")
	}

	c.log.Debug(c.ctx, "Data saved",
		log.String("key_id", fmt.Sprintf("%x", data.AuthKeyID)),
	)
	return nil
}

func (c *Client) onSession(cfg tg.Config, s mtproto.Session) error {
	sessionData := dcSessionFromMTProto(cfg.ThisDC, s)
	if c.isDroppedKey(sessionData) {
		// A connection that was still running on a key dropped after
		// ErrPFSDropKeysRequired reported its session: storing it would make
		// new connections reuse the dropped key.
		c.log.Debug(context.Background(), "Ignoring session with dropped auth key",
			log.Int("dc_id", sessionData.DC),
		)
		return nil
	}
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

// isDroppedKey reports whether data carries an auth key dropped for its DC
// after ErrPFSDropKeysRequired.
func (c *Client) isDroppedKey(data pool.Session) bool {
	if data.AuthKey.Zero() {
		return false
	}
	c.sessionsMux.Lock()
	defer c.sessionsMux.Unlock()
	dropped, ok := c.droppedKeys[data.DC]
	return ok && dropped == data.AuthKey.ID
}

func (c *Client) storeDCSess(target map[int]*pool.SyncSession, data pool.Session) {
	c.sessionsMux.Lock()
	if existing, ok := target[data.DC]; ok {
		if prev := existing.Load(); data.KeyCreatedAt == 0 && prev.AuthKey.ID == data.AuthKey.ID {
			// Same key reported by a connection that doesn't know its age.
			data.KeyCreatedAt = prev.KeyCreatedAt
		}
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
		DC:           dc,
		Salt:         s.Salt,
		AuthKey:      keyToStore,
		KeyCreatedAt: s.KeyCreatedAt,
	}
}
