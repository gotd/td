package mtproto

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/crypto"
)

// Session represents connection state.
type Session struct {
	ID   int64
	Key  crypto.AuthKey
	Salt int64
	// PermKey is non-zero only in PFS mode and is used for bindTempAuthKey.
	PermKey crypto.AuthKey
}

// Session returns current connection session info.
func (c *Conn) session() Session {
	c.updateSalt()

	c.sessionMux.RLock()
	defer c.sessionMux.RUnlock()
	return Session{
		Key:     c.authKey,
		Salt:    c.salt,
		ID:      c.sessionID,
		PermKey: c.permKey,
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
