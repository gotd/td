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

	if err := client.Connect(ctx); err != nil {
		panic(err)
	}

	if err := client.Close(); err != nil {
		panic(err)
	}
}
