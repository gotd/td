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

func sendStyledText(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	// This example creates a big message with different style lines
	// and sends to your Saved Messages folder.
	return client.Run(ctx, func(ctx context.Context) error {
		formats := []message.StyledTextOption{
			message.Plain("plaintext"), message.Plain("\n\n"),
			message.Mention("@durov"), message.Plain("\n\n"),
			message.Hashtag("#hashtag"), message.Plain("\n\n"),
			message.BotCommand("/command"), message.Plain("\n\n"),
			message.URL("https://google.org"), message.Plain("\n\n"),
			message.Email("example@example.org"), message.Plain("\n\n"),
			message.Bold("bold"), message.Plain("\n\n"),
			message.Italic("italic"), message.Plain("\n\n"),
			message.Underline("underline"), message.Plain("\n\n"),
			message.Strike("strike"), message.Plain("\n\n"),
			message.Code("fmt.Println(`Hello, World!`)"), message.Plain("\n\n"),
			message.Pre("fmt.Println(`Hello, World!`)", "Go"), message.Plain("\n\n"),
			message.TextURL("clickme", "https://google.com"), message.Plain("\n\n"),
			message.Phone("+71234567891"), message.Plain("\n\n"),
			message.Cashtag("$cashtag"), message.Plain("\n\n"),
			message.Blockquote("blockquote"), message.Plain("\n\n"),
			message.BankCard("5550111111111111"), message.Plain("\n\n"),
		}

		_, err := message.NewSender(tg.NewClient(client)).
			Self().StyledText(ctx, formats[0], formats[1:]...)
		return err
	})
}

func ExampleBuilder_StyledText() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := sendStyledText(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
