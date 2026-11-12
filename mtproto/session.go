package mtproto

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/crypto"
)

// Session represents connection state.
type Session struct {
	ID   int64
	Key  crypto.AuthKey
	Salt int64
	// PermKey is non-zero only in PFS mode and is used for bindTempAuthKey.
	PermKey crypto.AuthKey
	// KeyCreatedAt is the creation time of the permanent key (Key, or PermKey
	// in PFS mode), Unix seconds; zero if unknown.
	KeyCreatedAt int64
}

// Session returns current connection session info.
func (c *Conn) session() Session {
	c.updateSalt()

	c.sessionMux.RLock()
	defer c.sessionMux.RUnlock()
	return Session{
		Key:          c.authKey,
		Salt:         c.salt,
		ID:           c.sessionID,
		PermKey:      c.permKey,
		KeyCreatedAt: c.permKeyCreatedAt,
	}
}

// newSessionID sets session id to random value.
func (c *Conn) newSessionID() error {
	id, err := crypto.RandInt64(c.rand)
	if err != nil {
		return err
	}

	c.sessionMux.Lock()
	old := c.sessionID
	c.sessionID = id
	c.sessionMux.Unlock()

	c.traceSession(old, id)
	return nil
}

// reuseSession continues the server session saved in sessionState for key,
// if any.
func (c *Conn) reuseSession(ctx context.Context, key crypto.AuthKey) bool {
	if c.sessionState == nil || c.pfs {
		return false
	}
	id, n, ok := c.sessionState.load(key.ID)
	if !ok {
		return false
	}
	c.sessionMux.Lock()
	old := c.sessionID
	c.sessionID = id
	c.sessionMux.Unlock()
	c.reqMux.Lock()
	c.sentContentMessages = n
	c.reqMux.Unlock()
	c.sessionContinued = true
	c.log.Debug(ctx, "Continuing session", log.Int32("content_messages", n))
	c.traceSession(old, id)
	return true
}

// releaseSessionState saves the session of this connection for the next one.
func (c *Conn) releaseSessionState() {
	s := c.session()
	c.reqMux.Lock()
	n := c.sentContentMessages
	c.reqMux.Unlock()
	id := s.ID
	if c.pfs || s.Key.Zero() {
		// PFS: the temporary key and its session end with the connection.
		id = 0
	}
	select {
	case <-c.sessionFailed.Ready():
		// The next connection starts a new session.
		id = 0
	default:
	}
	c.sessionState.release(s.Key.ID, id, n)
}

// errSessionFailed stops Run when the server session can't be used any more,
// see failSession.
var errSessionFailed = errors.New("server session failed")

// failSession ends the server session after bad_msg_notification "msg_id
// too high": msg_ids already sent in it were ahead of the server time, and
// ids must keep increasing within a session, so corrected ids can only be
// used in a new one. The message id generator is reset to the corrected
// time and the connection is replaced by one in a new session; pending
// requests are sent again there. This is what TDLib does
// (SessionConnection: "Session will be closed", a new Session with new
// AuthData follows).
func (c *Conn) failSession() {
	if r, ok := c.messageID.(interface{ Reset() }); ok {
		r.Reset()
	}
	c.sessionFailed.Signal()
}

// errSessionReset is reported to requests that were pending in a session
// that is gone, so they are sent again with new msg_id and seqno.
var errSessionReset = errors.New("session reset")

// currentSessionID returns current session id.
func (c *Conn) currentSessionID() int64 {
	c.sessionMux.RLock()
	defer c.sessionMux.RUnlock()
	return c.sessionID
}

// resetSession clears the state bound to the previous session after the
// auth key and session were replaced in place: the new session starts with
// seqno 0, an empty received msg_id buffer and no cached future salts, and
// pending requests are sent again in it. Pending acks are dropped by ackLoop.
//
// This matches tdesktop (SessionPrivate::changeSessionId on a key change,
// followed by resendAll).
//
// Must be called with exchangeLock held for writing, so nothing is sent
// with the old msg_id and seqno in the new session.
func (c *Conn) resetSession() {
	c.reqMux.Lock()
	c.sentContentMessages = 0
	// Message ids are monotonic: everything allocated before this boundary
	// belongs to the old session and is refused by write.
	c.sessionStartMsgID.Store(c.newMessageID())
	c.reqMux.Unlock()

	if b, ok := c.messageIDBuf.(interface{ Reset() }); ok {
		b.Reset()
	}
	c.salts.Reset()
	if c.rpc != nil {
		c.rpc.NotifyErrorAll(errSessionReset)
	}
}
