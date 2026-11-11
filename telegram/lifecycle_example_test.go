package telegram_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
)

// Bot without Run: Start, use the client, then Wait until it stops.
func ExampleClient_Start() {
	if err := func() error {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		client, err := telegram.ClientFromEnvironment(telegram.Options{})
		if err != nil {
			return err
		}

		// Bound the startup: the client reconnects until ready otherwise.
		startCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		// Stop the client on interrupt. Without WithLifetime it runs until
		// Close or a fatal error: startCtx bounds the startup only.
		if err := client.Start(startCtx, telegram.WithLifetime(ctx)); err != nil {
			return err
		}
		defer func() { _ = client.Close() }()

		if _, err := client.Auth().Bot(ctx, os.Getenv("BOT_TOKEN")); err != nil {
			return err
		}
		sender := message.NewSender(client.API())
		if _, err := sender.Resolve("@durov").Text(ctx, "Hello!"); err != nil {
			return err
		}

		// Block until interrupt or a fatal error, e.g. *telegram.AuthLostError.
		return client.Wait(context.Background())
	}(); err != nil {
		fmt.Println(err)
	}
}
