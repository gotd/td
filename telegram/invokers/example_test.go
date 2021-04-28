package invokers_test

import (
	"context"
	"time"

	"go.uber.org/ratelimit"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/telegram/invokers"
	"github.com/gotd/td/tg"
)

func ExampleRateLimiter() {
	var invoker tg.Invoker // e.g. *telegram.Client

	limiter := invokers.NewRateLimiter(invoker, ratelimit.New(1,
		ratelimit.Per(100*time.Millisecond),
	))

	tg.NewClient(limiter)
}

func ExampleWaiter() {
	var invoker tg.Invoker // e.g. *telegram.Client

	waiter := invokers.NewWaiter(invoker).
		WithMaxWait(5 * time.Minute).
		WithMaxRetries(3)

	tg.NewClient(waiter)
}

func ExampleWaitScheduler() {
	var invoker tg.Invoker // e.g. *telegram.Client

	waiter := invokers.NewWaitScheduler(invoker).
		WithWaitLimit(300).
		WithRetryLimit(3)

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return waiter.Run(ctx)
	})
	g.Go(func() error {
		// do something with waiter invoker or client
		// e.g. raw := tg.NewClient(waiter)
		return nil
	})
	err := g.Wait()
	if err != nil {
		panic(err)
	}
}
