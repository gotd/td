package transport_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/transport"

	"github.com/stretchr/testify/require"
)

func TestMTProxy(t *testing.T) {
	_, err := transport.MTProxy(nil, 0, nil)
	require.Error(t, err)
}

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

	trp, err := transport.MTProxy(nil, 2, secret)
	if err != nil {
		panic(err)
	}

	// Creating connection.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	client := telegram.NewClient(1, "appHash", telegram.Options{
		Addr:      addr,
		Transport: trp,
	})

	go func() { _ = client.Run(ctx) }()
}
