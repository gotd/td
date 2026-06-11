package pool

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/atomic"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/tdsync"
)

// ErrConnDead means that connection is closed and can't be used anymore.
var ErrConnDead = errors.New("connection dead")

// errRetryableOnNewConn reports whether request failed because connection
// died before the request was processed by the server (request was not sent,
// or sent but not acknowledged), so it is safe to retry the request on a new
// connection.
func errRetryableOnNewConn(err error) bool {
	return errors.Is(err, ErrConnDead) || errors.Is(err, rpc.ErrEngineClosed)
}

// Conn represents Telegram MTProto connection.
type Conn interface {
	Run(ctx context.Context) error
	Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error
	Ping(ctx context.Context) error
	Ready() <-chan struct{}
}

type poolConn struct {
	Conn
	id      int64 // immutable
	dc      *DC   // immutable
	deleted *atomic.Bool
	dead    *tdsync.Ready
}

func (p *poolConn) Dead() <-chan struct{} {
	return p.dead.Ready()
}
