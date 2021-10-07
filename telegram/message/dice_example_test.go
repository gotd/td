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

func sendDice(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		sender := message.NewSender(tg.NewClient(client))

		// Sends dice "🎲" to the @durovschat.
		if _, err := sender.Resolve("@durovschat").Dice(ctx); err != nil {
			return err
		}

		// Sends darts "🎯" to the @durovschat.
		if _, err := sender.Resolve("https://t.me/durovschat").Darts(ctx); err != nil {
			return err
		}

		return nil
	})
}

func ExampleMediaDice() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := sendDice(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
