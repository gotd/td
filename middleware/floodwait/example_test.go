package floodwait_test

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/middleware/floodwait"
	"github.com/gotd/td/tg"
)

func ExampleSimpleWaiter() {
	var invoker tg.Invoker // e.g. *telegram.Client

	waiter := floodwait.NewSimpleWaiter(invoker).
		WithMaxWait(5 * time.Minute).
		WithMaxRetries(3)

	// Do something with waiter invoker.
	// E.g. create a new RPC client.
	tg.NewClient(waiter)
}

func ExampleWaiter() {
	var invoker tg.Invoker // e.g. *telegram.Client

	waiter := floodwait.NewWaiter(invoker).
		WithMaxWait(5 * time.Minute).
		WithMaxRetries(3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return waiter.Run(ctx)
	})
	g.Go(func() error {
		// Cancel context for waiter goroutine even
		// if we return nil error.
		defer cancel()

		// Do something with waiter invoker.
		// E.g. create a new RPC client.
		tg.NewClient(waiter)

		return nil
	})
	if err := g.Wait(); err != nil {
		panic(err)
	}
}
