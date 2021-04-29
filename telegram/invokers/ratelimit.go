package invokers

import (
	"context"

	"go.uber.org/ratelimit"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// RateLimiter is a tg.Invoker that throttles RPC calls on underlying invoker.
type RateLimiter struct {
	next tg.Invoker
	rlim ratelimit.Limiter
}

// NewRateLimiter returns a new invoker rate limiter using rlim.
func NewRateLimiter(invoker tg.Invoker, rlim ratelimit.Limiter) *RateLimiter {
	return &RateLimiter{
		next: invoker,
		rlim: rlim,
	}
}

// InvokeRaw implements tg.Invoker.
func (l *RateLimiter) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	// TODO(tie) support proper context cancellation for rate limits.
	//
	// See https://github.com/uber-go/ratelimit/pull/11#issuecomment-424912370
	//
	if err := ctx.Err(); err != nil {
		return err
	}
	l.rlim.Take()
	return l.next.InvokeRaw(ctx, input, output)
}
