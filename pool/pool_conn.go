package pool

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/atomic"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/transport"
)

// ErrConnDead means that connection is closed and can't be used anymore.
var ErrConnDead = errors.New("connection dead")

// errRetryableOnNewConn reports whether request failed because connection
// died before the request was processed by the server (request was not sent,
// or sent but not acknowledged), so it is safe to retry the request on a new
// connection.
//
// A transport write failure also qualifies — the frame was either not sent at
// all or sent partially, and the server discards partial frames — but only when
// the caller opted in through DCOptions.RetryOnWriteFailed, since retrying it
// hides the error from a caller that acts on it itself.
func errRetryableOnNewConn(err error, retryOnWriteFailed bool) bool {
	if errors.Is(err, ErrConnDead) || errors.Is(err, rpc.ErrEngineClosed) {
		return true
	}
	return retryOnWriteFailed && errors.Is(err, transport.ErrWriteFailed)
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
