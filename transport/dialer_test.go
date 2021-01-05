package transport_test

import (
	"context"
	"fmt"
	"net"
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

// ExampleDialer for socks5. Methods form https://stackoverflow.com/questions/59456936/socks5-proxy-client-with-context-support
func ExampleDialer() {
	// No error would be return.
	sock5, _ := proxy.SOCKS5("tcp", "IP:PORT", &proxy.Auth{
		User:     "YOURUSERNAME",
		Password: "YOURPASSWORD",
	}, proxy.Direct)

	dc := sock5.(interface {
		DialContext(ctx context.Context, network, addr string) (net.Conn, error)
	})

	// Creating connection.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	client := telegram.NewClient(1, "appHash", telegram.Options{
		Transport: transport.Intermediate(transport.DialFunc(dc.DialContext)),
	})

	_ = client.Run(ctx, func(ctx context.Context) error {
		fmt.Println("Started")
		return nil
	})
}
