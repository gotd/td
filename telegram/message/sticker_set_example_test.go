package message_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

func sendSticker(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		sender := message.NewSender(tg.NewClient(client))

		// Sends the first favorite sticker to the @durovschat.
		if _, err := sender.Resolve("https://t.me/durovschat").
			Sticker(message.FavedStickers()).
			First(ctx); err != nil {
			return err
		}

		// Sends a recently used sticker by its emoji to your Saved Messages.
		if _, err := sender.Self().
			Sticker(message.RecentStickers()).
			ByEmoji(ctx, "😎"); err != nil {
			return err
		}

		// Sends the sticker at index 0 of a sticker set by its short name.
		if _, err := sender.Self().
			Sticker(message.StickerSetName("AnimatedEmojies")).
			ByIndex(ctx, 0); err != nil {
			return err
		}

		return nil
	})
}

func ExampleStickerSetBuilder() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := sendSticker(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
