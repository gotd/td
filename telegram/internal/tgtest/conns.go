package tgtest

import (
	"net"
	"sync"

	"github.com/gotd/td/internal/crypto"
)

type conns struct {
	mux sync.Mutex
	m   map[crypto.AuthKey]net.Conn
}

func newConns() *conns {
	return &conns{m: map[crypto.AuthKey]net.Conn{}}
}

func (c *conns) add(key crypto.AuthKey, conn net.Conn) {
	c.mux.Lock()
	c.m[key] = conn
	c.mux.Unlock()
}

func (c *conns) get(key crypto.AuthKey) (conn net.Conn) {
	c.mux.Lock()
	conn = c.m[key]
	c.mux.Unlock()

	return
}

func (c *conns) delete(key crypto.AuthKey) {
	c.mux.Lock()
	delete(c.m, key)
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
