package unpack_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/telegram/message/unpack"
	"github.com/nnqq/td/tg"
)

func unpackMessage(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		sender := message.NewSender(tg.NewClient(client))

		msg, err := unpack.Message(sender.Resolve("@durovschat").Dice(ctx))
		// Sends dice "🎲" to the @durovschat.
		if err != nil {
			return err
		}

		fmt.Println("Sent message ID:", msg.ID)

		return nil
	})
}

func ExampleMessage() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := unpackMessage(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
