package tgtest

import (
	"sync"

	"go.uber.org/atomic"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/transport"
)

// connection is the transport connection of a session. A client may continue
// a session on a new transport connection: the server then answers on the
// latest one, like Telegram.
type connection struct {
	mux  sync.Mutex
	conn transport.Conn
	sent atomic.Bool
}

func (conn *connection) sentCreated() bool {
	return conn.sent.Swap(true)
}

func (conn *connection) get() transport.Conn {
	conn.mux.Lock()
	defer conn.mux.Unlock()
	return conn.conn
}

func (conn *connection) set(c transport.Conn) {
	conn.mux.Lock()
	conn.conn = c
	conn.mux.Unlock()
}

// Close closes the current transport connection.
func (conn *connection) Close() error {
	if c := conn.get(); c != nil {
		return c.Close()
	}
	return nil
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
		if v.get() != tConn {
			// The session continues on a new connection.
			v.set(tConn)
		}
		return v
	}

	conn := &connection{
		conn: tConn,
	}
	c.conns[key] = conn
	return conn
}

func (c *users) forgetSession(key int64) {
	c.connsMux.Lock()
	if conn, ok := c.conns[key]; ok {
		conn.sent.Store(false)
	}
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
