package tgtest

import (
	"sync"

	"go.uber.org/atomic"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/transport"
)

type connection struct {
	transport.Conn
	sent atomic.Bool
}

func (conn *connection) sentCreated() bool {
	return conn.sent.Swap(true)
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

func (c *users) createConnection(key int64, tConn transport.Conn) *connection {
	c.connsMux.Lock()
	defer c.connsMux.Unlock()

	if v, ok := c.conns[key]; ok {
		return v
	}

	conn := &connection{
		Conn: tConn,
	}
	c.conns[key] = conn
	return conn
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
