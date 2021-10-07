package message_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/telegram/message/styling"
	"github.com/nnqq/td/tg"
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
			styling.Plain("plaintext"), styling.Plain("\n\n"),
			styling.Mention("@durov"), styling.Plain("\n\n"),
			styling.Hashtag("#hashtag"), styling.Plain("\n\n"),
			styling.BotCommand("/command"), styling.Plain("\n\n"),
			styling.URL("https://google.org"), styling.Plain("\n\n"),
			styling.Email("example@example.org"), styling.Plain("\n\n"),
			styling.Bold("bold"), styling.Plain("\n\n"),
			styling.Italic("italic"), styling.Plain("\n\n"),
			styling.Underline("underline"), styling.Plain("\n\n"),
			styling.Strike("strike"), styling.Plain("\n\n"),
			styling.Code("fmt.Println(`Hello, World!`)"), styling.Plain("\n\n"),
			styling.PreLang("fmt.Println(`Hello, World!`)", "Go"), styling.Plain("\n\n"),
			styling.TextURL("clickme", "https://google.com"), styling.Plain("\n\n"),
			styling.Phone("+71234567891"), styling.Plain("\n\n"),
			styling.Cashtag("$CASHTAG"), styling.Plain("\n\n"),
			styling.Blockquote("blockquote"), styling.Plain("\n\n"),
			styling.BankCard("5550111111111111"), styling.Plain("\n\n"),
		}

		_, err := message.NewSender(tg.NewClient(client)).
			Self().StyledText(ctx, formats...)
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
