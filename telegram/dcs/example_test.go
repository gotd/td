package dcs_test

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/net/proxy"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
)

func ExampleDialFunc() {
	// Creating connection.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	client := telegram.NewClient(1, "appHash", telegram.Options{
		Resolver: dcs.PlainResolver(dcs.PlainOptions{DialContext: proxy.Dial}),
	})

	_ = client.Run(ctx, func(ctx context.Context) error {
		fmt.Println("Started")
		return nil
	})
}

// ExampleDialer for socks5. Methods form https://stackoverflow.com/questions/59456936/socks5-proxy-client-with-context-support
func ExampleDialer() {
	// No error would be return.
	sock5, _ := proxy.SOCKS5("tcp", "IP:PORT", &proxy.Auth{
		User:     "YOURUSERNAME",
		Password: "YOURPASSWORD",
	}, proxy.Direct)
	dc := sock5.(proxy.ContextDialer)

	// Creating connection.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	client := telegram.NewClient(1, "appHash", telegram.Options{
		Resolver: dcs.PlainResolver(dcs.PlainOptions{DialContext: dc.DialContext}),
	})

	_ = client.Run(ctx, func(ctx context.Context) error {
		fmt.Println("Started")
		return nil
	})
}
