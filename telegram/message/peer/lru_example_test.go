package peer_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/telegram/message/peer"
	"github.com/nnqq/td/tg"
)

func resolveLRU(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		raw := tg.NewClient(client)
		resolver := peer.NewLRUResolver(peer.Plain(raw), 16).WithExpiration(time.Minute)
		sender := message.NewSender(raw).WithResolver(resolver)

		// "durovschat" will be resolved by Plain resolver.
		if _, err := sender.Resolve("@durovschat").Dice(ctx); err != nil {
			return err
		}

		// "durovschat" will be resolved by cache.
		if _, err := sender.Resolve("https://t.me/durovschat").Darts(ctx); err != nil {
			return err
		}

		// Evict and delete record.
		resolver.Evict("durovschat")

		return nil
	})
}

func ExampleLRUResolver_cache() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := resolveLRU(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
