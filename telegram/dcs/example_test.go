package dcs_test

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/net/proxy"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/dcs"
)

func ExampleDialFunc() {
	// Dial using proxy from environment.

	// Creating connection.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	client := telegram.NewClient(1, "appHash", telegram.Options{
		Resolver: dcs.Plain(dcs.PlainOptions{Dial: proxy.Dial}),
	})

	_ = client.Run(ctx, func(ctx context.Context) error {
		fmt.Println("Started")
		return nil
	})
}

func ExampleDialFunc_dialer() {
	// Dial using SOCKS5 proxy.

	sock5, _ := proxy.SOCKS5("tcp", "IP:PORT", &proxy.Auth{
		User:     "YOURUSERNAME",
		Password: "YOURPASSWORD",
	}, proxy.Direct)
	dc := sock5.(proxy.ContextDialer)

	// Creating connection.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	client := telegram.NewClient(1, "appHash", telegram.Options{
		Resolver: dcs.Plain(dcs.PlainOptions{
			Dial: dc.DialContext,
		}),
	})

	_ = client.Run(ctx, func(ctx context.Context) error {
		fmt.Println("Started")
		return nil
	})
}
