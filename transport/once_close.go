package transport

import (
	"net"
	"sync"
)

// onceCloseListener wraps a net.Listener, protecting it from
// multiple Close calls.
type onceCloseListener struct {
	net.Listener
	once sync.Once
	err  error
}

func (o *onceCloseListener) Close() error {
	o.once.Do(func() {
		o.err = o.Listener.Close()
	})
	return o.err
}
