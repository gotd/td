package transport_test

import (
	"context"
	"fmt"
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

	_ = client.Run(ctx, func(ctx context.Context) error {
		fmt.Println("Started")
		return nil
	})
}
