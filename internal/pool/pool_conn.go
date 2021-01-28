package pool

import (
	"go.uber.org/atomic"
)

type poolConn struct {
	*conn
	dc   *DC // immutable
	dead atomic.Bool
}
