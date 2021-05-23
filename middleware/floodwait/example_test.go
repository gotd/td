package floodwait_test

import (
	"time"

	"github.com/gotd/td/middleware/floodwait"
	"github.com/gotd/td/tg"
)

func ExampleWaiter() {
	var invoker tg.Invoker // e.g. *telegram.Client

	waiter := floodwait.NewWaiter(invoker).
		WithMaxWait(5 * time.Minute).
		WithMaxRetries(3)

	// Do something with waiter invoker.
	// E.g. create a new RPC client.
	tg.NewClient(waiter)
}
