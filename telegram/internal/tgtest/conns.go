package tgtest

import (
	"sync"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/transport"
)

type conns struct {
	mux sync.Mutex
	m   map[crypto.AuthKey]transport.Connection
}

func newConns() *conns {
	return &conns{m: map[crypto.AuthKey]transport.Connection{}}
}

func (c *conns) add(key Session, conn transport.Connection) {
	c.mux.Lock()
	c.m[key.Key] = conn
	c.mux.Unlock()
}

func (c *conns) get(key Session) (conn transport.Connection, ok bool) {
	c.mux.Lock()
	conn, ok = c.m[key.Key]
	c.mux.Unlock()

	return
}

func (c *conns) delete(key Session) {
	c.mux.Lock()
	delete(c.m, key.Key)
	c.mux.Unlock()
}

func (c *conns) Close() error {
	c.mux.Lock()
	for _, conn := range c.m {
		_ = conn.Close()
	}
	c.mux.Unlock()

	return nil
}
