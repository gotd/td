package tgtest

import (
	"sync"
	"sync/atomic"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/transport"
)

type connection struct {
	transport.Conn
	sent uint64
}

func (conn *connection) didSentCreated() bool {
	return atomic.LoadUint64(&conn.sent) >= 1
}

func (conn *connection) sentCreated() {
	atomic.AddUint64(&conn.sent, 1)
}

type users struct {
	sessions    map[[8]byte]crypto.AuthKeyWithID
	sessionsMux sync.Mutex

	conns    map[[8]byte]*connection
	connsMux sync.Mutex
}

func newUsers() *users {
	return &users{
		conns:    map[[8]byte]*connection{},
		sessions: map[[8]byte]crypto.AuthKeyWithID{},
	}
}

func (c *users) createSession(key crypto.AuthKeyWithID, conn *connection) {
	c.addConnection(key, conn)
	c.addSession(key)
}

func (c *users) addConnection(key crypto.AuthKeyWithID, conn *connection) {
	c.connsMux.Lock()
	c.conns[key.AuthKeyID] = conn
	c.connsMux.Unlock()
}

func (c *users) getConnection(key crypto.AuthKeyWithID) (conn *connection, ok bool) {
	c.connsMux.Lock()
	conn, ok = c.conns[key.AuthKeyID]
	c.connsMux.Unlock()

	return
}

func (c *users) deleteConnection(key crypto.AuthKeyWithID) {
	c.connsMux.Lock()
	conn := c.conns[key.AuthKeyID]
	if conn != nil {
		_ = conn.Close()
	}
	delete(c.conns, key.AuthKeyID)
	c.connsMux.Unlock()
}

func (c *users) addSession(key crypto.AuthKeyWithID) {
	c.sessionsMux.Lock()
	c.sessions[key.AuthKeyID] = key
	c.sessionsMux.Unlock()
}

func (c *users) getSession(k [8]byte) (s crypto.AuthKeyWithID, ok bool) {
	c.connsMux.Lock()
	s, ok = c.sessions[k]
	c.connsMux.Unlock()

	return
}

func (c *users) Close() error {
	c.connsMux.Lock()
	for _, conn := range c.conns {
		_ = conn.Close()
	}
	c.connsMux.Unlock()

	return nil
}
