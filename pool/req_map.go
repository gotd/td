package pool

import (
	"sync"

	"go.uber.org/atomic"
)

type reqKey int64

type reqMap struct {
	m   map[reqKey]chan *poolConn
	mux sync.Mutex
	_   [4]byte

	nextRequest atomic.Int64
}

func newReqMap() *reqMap {
	return &reqMap{
		m: map[reqKey]chan *poolConn{},
	}
}

func (r *reqMap) request() (key reqKey, ch chan *poolConn) {
	key = reqKey(r.nextRequest.Inc())
	ch = make(chan *poolConn, 1)

	r.mux.Lock()
	r.m[key] = ch
	r.mux.Unlock()
	return key, ch
}

func (r *reqMap) transfer(c *poolConn) bool {
	r.mux.Lock()
	if len(r.m) < 1 { // no requests
		r.mux.Unlock()
		return false
	}

	var ch chan *poolConn
	var k reqKey
	for k, ch = range r.m { // Get one from map.
		break
	}
	delete(r.m, k) // Remove from pending requests.
	r.mux.Unlock()

	if ch == nil {
		panic("unreachable: channel can't be nil due to map not empty")
	}

	ch <- c
	close(ch)
	return true
}

func (r *reqMap) delete(key reqKey) {
	r.mux.Lock()
	delete(r.m, key)
	r.mux.Unlock()
}
