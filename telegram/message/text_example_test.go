package message_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/tg"
)

func sendText(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	// This example creates a plaing message and sends it
	// to your Saved Message folder.
	return client.Run(ctx, func(ctx context.Context) error {
		_, err := message.NewSender(tg.NewClient(client)).Self().Text(ctx, "Hi!")
		return err
	})
}

func ExampleBuilder_Text() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := sendText(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
