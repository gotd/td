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

// users contains all server connections and sessions.
type users struct {
	sessions    map[[8]byte]crypto.AuthKey
	sessionsMux sync.Mutex

	conns    map[int64]*connection
	connsMux sync.Mutex
}

func newUsers() *users {
	return &users{
		conns:    map[int64]*connection{},
		sessions: map[[8]byte]crypto.AuthKey{},
	}
}

func (c *users) addConnection(key int64, conn *connection) {
	c.connsMux.Lock()
	c.conns[key] = conn
	c.connsMux.Unlock()
}

func (c *users) getConnection(key int64) (conn *connection, ok bool) {
	c.connsMux.Lock()
	conn, ok = c.conns[key]
	c.connsMux.Unlock()

	return
}

func (c *users) deleteConnection(key int64) {
	c.connsMux.Lock()
	conn := c.conns[key]
	if conn != nil {
		_ = conn.Close()
	}
	delete(c.conns, key)
	c.connsMux.Unlock()
}

func (c *users) addSession(key crypto.AuthKey) {
	c.sessionsMux.Lock()
	c.sessions[key.ID] = key
	c.sessionsMux.Unlock()
}

func (c *users) getSession(k [8]byte) (s crypto.AuthKey, ok bool) {
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
