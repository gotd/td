package invokers_test

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"
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
		WithMaxWait(5 * time.Minute).
		WithMaxRetries(3)

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
