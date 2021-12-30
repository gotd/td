package message_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
)

func sendText(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		s := message.NewSender(client.API())

		// Sends message to your Saved Message folder.
		if _, err := s.Self().Text(ctx, "Hi!"); err != nil {
			return err
		}

		// Resolves @my_channel and gets tg.InputPeerClass.
		p, err := s.Resolve("my_channel").AsInputPeer(ctx)
		if err != nil {
			return err
		}

		// Replies to message ID = 1 in @gotd_en as @my_channel with spoiler message "spoiler".
		if _, err := s.Resolve("gotd_en").
			SendAs(p).Reply(1).
			StyledText(ctx, styling.Spoiler("spoiler")); err != nil {
			return err
		}

		// Sends message to channel
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
