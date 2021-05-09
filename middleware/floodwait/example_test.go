package floodwait_test

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/middleware/floodwait"
	"github.com/gotd/td/tg"
)

func ExampleWaitTimer() {
	var invoker tg.Invoker // e.g. *telegram.Client

	waiter := floodwait.NewWaitTimer(invoker).
		WithMaxWait(5 * time.Minute).
		WithMaxRetries(3)

	tg.NewClient(waiter)
}

func ExampleWaitScheduler() {
	var invoker tg.Invoker // e.g. *telegram.Client

	waiter := floodwait.NewWaitScheduler(invoker).
		WithMaxWait(5 * time.Minute).
		WithMaxRetries(3)

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return waiter.Run(ctx)
	})
	g.Go(func() error {
		defer cancel() // always cancel context for waiter goroutine

		// do something with waiter invoker or client
		// e.g. raw := tg.NewClient(waiter)
		return nil
	})
	err := g.Wait()
	if err != nil {
		panic(err)
	}
}
