package middleware

import (
	"github.com/gotd/td/tg"
)

// Middleware is a tg.Invoker middleware constructor.
type Middleware func(invoker tg.Invoker) tg.Invoker

// Chain returns a Middleware that chains multiple Middleware constructor
// calls.
func Chain(middlewares ...Middleware) Middleware {
	return func(invoker tg.Invoker) tg.Invoker {
		for _, f := range middlewares {
			invoker = f(invoker)
		}
		return invoker
	}
}
