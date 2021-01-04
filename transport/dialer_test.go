package transport_test

import (
	"context"
	"time"

	"golang.org/x/net/proxy"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/transport"
)

func ExampleDialFunc() {
	trp := transport.Intermediate(transport.DialFunc(proxy.Dial))

	// Creating connection.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	client := telegram.NewClient(1, "appHash", telegram.Options{
		Transport: trp,
	})

	go func() { _ = client.Run(ctx) }()
}
