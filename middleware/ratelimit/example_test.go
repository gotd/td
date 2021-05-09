package ratelimit_test

import (
	"time"

	"golang.org/x/time/rate"

	"github.com/gotd/td/middleware/ratelimit"
	"github.com/gotd/td/tg"
)

func ExampleRateLimiter() {
	var invoker tg.Invoker // e.g. *telegram.Client

	limiter := ratelimit.NewRateLimiter(invoker,
		rate.NewLimiter(rate.Every(100*time.Millisecond), 1),
	)

	tg.NewClient(limiter)
}
