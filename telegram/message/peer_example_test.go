package message_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/telegram/message/peer"
	"github.com/nnqq/td/tg"
)

func resolve(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		sender := message.NewSender(tg.NewClient(client))

		// Resolve and return input peer for @misato.
		_, err := sender.Resolve("misato").AsInputPeer(ctx)
		if err != nil {
			return err
		}

		// Resolve and join channel @seele.
		// If @seele is a user, not channel, error would be returned.
		_, err = sender.Resolve("seele", peer.OnlyChannel).Join(ctx)
		if err != nil {
			return err
		}

		return nil
	})
}

func ExampleSender_Resolve() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := resolve(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
