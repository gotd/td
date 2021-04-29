package invokers_test

import (
	"time"

	"golang.org/x/time/rate"

	"github.com/gotd/td/telegram/invokers"
	"github.com/gotd/td/tg"
)

func ExampleRateLimiter() {
	var invoker tg.Invoker // e.g. *telegram.Client

	limiter := invokers.NewRateLimiter(invoker,
		rate.NewLimiter(rate.Every(100*time.Millisecond), 1),
	)

	tg.NewClient(limiter)
}
