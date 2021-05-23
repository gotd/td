package floodwait

import (
	"github.com/gotd/td/middleware"
	"github.com/gotd/td/tg"
)

// MiddlewareOption configures new Waiter in middleware constructor.
type MiddlewareOption func(w *Waiter) *Waiter

// Middleware returns a new Waiter middleware constructor.
func Middleware(opts ...MiddlewareOption) middleware.Middleware {
	return func(invoker tg.Invoker) tg.Invoker {
		waiter := NewWaiter(invoker)
		for _, f := range opts {
			waiter = f(waiter)
		}
		return waiter
	}
}
