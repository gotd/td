package transport_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/transport"
)

func ExampleMTProxy() {
	addr, ok := os.LookupEnv("MTPROXY_ADDR")
	if !ok {
		fmt.Println("MTPROXY_ADDR is not set")
		return
	}

	secret, err := hex.DecodeString(os.Getenv("MTPROXY_SECRET"))
	if err != nil {
		panic(err)
	}

	trp, err := transport.MTProxy(nil, addr, secret)
	if err != nil {
		panic(err)
	}

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
