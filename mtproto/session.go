package mtproto

import "github.com/gotd/td/crypto"

// Session represents connection state.
type Session struct {
	ID   int64
	Key  crypto.AuthKey
	Salt int64
}

// Session returns current connection session info.
func (c *Conn) session() Session {
	c.updateSalt()

	c.sessionMux.RLock()
	defer c.sessionMux.RUnlock()
	return Session{
		Key:  c.authKey,
		Salt: c.salt,
		ID:   c.sessionID,
	}
}

// newSessionID sets session id to random value.
func (c *Conn) newSessionID() error {
	id, err := crypto.RandInt64(c.rand)
	if err != nil {
		return err
	}

	c.sessionMux.Lock()
	defer c.sessionMux.Unlock()
	c.sessionID = id

	return nil
}
