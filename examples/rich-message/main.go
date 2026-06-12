// Binary rich-message sends a structured "rich" message to Saved Messages
// using the telegram/message/rich helper.
//
// A rich message carries structured content (titles, headings, paragraphs,
// lists, dividers, ...) built from page blocks, instead of a flat string.
// The rich package provides constructors for the blocks and the inline rich
// text, and message.Builder.RichMessage sends the assembled message.
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/rich"
)

func run(ctx context.Context) error {
	log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel))
	defer func() { _ = log.Sync() }()

	// Initializing client from environment.
	// Available environment variables:
	// 	APP_ID:         app_id of Telegram app.
	// 	APP_HASH:       app_hash of Telegram app.
	// 	SESSION_FILE:   path to session file
	// 	SESSION_DIR:    path to session directory, if SESSION_FILE is not set
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: logzap.New(log),
	})
	if err != nil {
		return err
	}

	// Reading phone, code and 2FA password from terminal when no session.
	flow := auth.NewFlow(examples.Terminal{}, auth.SendCodeOptions{})

	return client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return errors.Wrap(err, "auth")
		}

		// Assemble a rich message from page blocks and inline rich text.
		msg := rich.New(
			rich.Title(rich.Plain("gotd rich message")),
			rich.Header(rich.Plain("Built with message/rich")),
			rich.Paragraph(rich.Concat(
				rich.Plain("This message mixes "),
				rich.Bold(rich.Plain("bold")),
				rich.Plain(", "),
				rich.Italic(rich.Plain("italic")),
				rich.Plain(" and "),
				rich.Fixed(rich.Plain("monospace")),
				rich.Plain(" text."),
			)),
			rich.List(
				rich.ListItem(rich.Plain("First item")),
				rich.ListItem(rich.Plain("Second item")),
			),
			rich.Divider(),
			rich.Footer(rich.Plain("Sent by the gotd rich-message example.")),
		).Input()

		sender := message.NewSender(client.API())
		if _, err := sender.Self().RichMessage(ctx, msg); err != nil {
			return errors.Wrap(err, "send rich message")
		}
		log.Info("Rich message sent to Saved Messages")
		return nil
	})
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}
