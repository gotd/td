package pool

import (
	"sync"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mtproto"
)

// Session represents DC session.
type Session struct {
	DC      int
	AuthKey crypto.AuthKey
	Salt    int64
}

// SyncSession is synchronization helper for Session.
type SyncSession struct {
	data Session
	mux  sync.RWMutex
}

// NewSyncSession creates new SyncSession.
func NewSyncSession(data Session) *SyncSession {
	return &SyncSession{
		data: data,
	}
}

// Store saves given Session.
func (s *SyncSession) Store(data Session) {
	s.mux.Lock()
	s.data = data
	s.mux.Unlock()
}

// Migrate changes current DC and its addr, zeroes AuthKey and Salt.
func (s *SyncSession) Migrate(dc int) {
	s.mux.Lock()
	s.data.DC = dc
	s.data.AuthKey = crypto.AuthKey{}
	s.data.Salt = 0
	s.mux.Unlock()
}

// Options fills Key and Salt field of given Options using stored session and returns it.
func (s *SyncSession) Options(opts mtproto.Options) (mtproto.Options, Session) {
	s.mux.RLock()
	data := s.data
	s.mux.RUnlock()

	if opts.EnablePFS {
		// Stored key in pool/session remains backward-compatible single "AuthKey".
		// In PFS mode this persisted key is treated as permanent key, while
		// temporary key is always generated per runtime connection.
		opts.PermKey = data.AuthKey
		opts.Key = crypto.AuthKey{}
	} else {
		opts.Key = data.AuthKey
	}
	opts.Salt = data.Salt
	return opts, data
}

// Load gets session and returns it.
func (s *SyncSession) Load() (data Session) {
	s.mux.RLock()
	data = s.data
	s.mux.RUnlock()

	return
}
