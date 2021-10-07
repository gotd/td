// Binary bot-echo implements basic example for bot.
package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/nnqq/td/examples"
	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/tg"
)

func main() {
	// Environment variables:
	//	BOT_TOKEN:     token from BotFather
	// 	APP_ID:        app_id of Telegram app.
	// 	APP_HASH:      app_hash of Telegram app.
	// 	SESSION_FILE:  path to session file
	// 	SESSION_DIR:   path to session directory, if SESSION_FILE is not set
	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		// Dispatcher handles incoming updates.
		dispatcher := tg.NewUpdateDispatcher()
		opts := telegram.Options{
			Logger:        log,
			UpdateHandler: dispatcher,
		}
		return telegram.BotFromEnvironment(ctx, opts, func(ctx context.Context, client *telegram.Client) error {
			// Raw MTProto API client, allows making raw RPC calls.
			api := tg.NewClient(client)

			// Helper for sending messages.
			sender := message.NewSender(api)

			// Setting up handler for incoming message.
			dispatcher.OnNewMessage(func(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage) error {
				m, ok := u.Message.(*tg.Message)
				if !ok || m.Out {
					// Outgoing message, not interesting.
					return nil
				}

				// Sending reply.
				_, err := sender.Reply(entities, u).Text(ctx, m.Message)
				return err
			})
			return nil
		}, telegram.RunUntilCanceled)
	})
}
