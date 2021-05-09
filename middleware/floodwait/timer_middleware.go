package floodwait

import (
	"github.com/gotd/td/middleware"
	"github.com/gotd/td/tg"
)

// TimerMiddlewareOption configures new WaitTimer in middleware constructor.
type TimerMiddlewareOption func(w *WaitTimer) *WaitTimer

// TimerMiddleware returns a new WaitTimer middleware constructor.
func TimerMiddleware(opts ...TimerMiddlewareOption) middleware.Middleware {
	return func(invoker tg.Invoker) tg.Invoker {
		waiter := NewWaitTimer(invoker)
		for _, f := range opts {
			waiter = f(waiter)
		}
		return waiter
	}
}
