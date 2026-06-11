package telegram_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gotd/td/telegram"
)

func ExampleOptions_connectionState() {
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatal("APP_ID not set or invalid")
	}
	appHash := os.Getenv("APP_HASH")

	client := telegram.NewClient(appID, appHash, telegram.Options{
		OnConnectionState: func(state telegram.ConnectionState) {
			// Called on primary connection state change: connecting, ready,
			// disconnected. Callback must not block.
			fmt.Println("connection state:", state)
		},
	})
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// Connection state transitions are reported while client is running,
		// including automatic reconnects.
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}
