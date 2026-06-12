// Binary bot-inline implements an inline bot that answers inline queries
// (@your_bot some text) with article results.
//
// It demonstrates the telegram/message/inline helper for building inline
// query results and the telegram/message/styling helper for styled message
// content. Enable inline mode for your bot via @BotFather (/setinline) first.
package main

import (
	"context"
	"crypto/rand"

	"go.uber.org/zap"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message/inline"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
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
			Logger:        logzap.New(log),
			UpdateHandler: dispatcher,
		}
		return telegram.BotFromEnvironment(ctx, opts, func(ctx context.Context, client *telegram.Client) error {
			api := tg.NewClient(client)

			// Handle inline queries (@your_bot <query>).
			dispatcher.OnBotInlineQuery(func(ctx context.Context, e tg.Entities, u *tg.UpdateBotInlineQuery) error {
				log.Info("Inline query", zap.String("query", u.Query))

				// inline.New builds an answer for the given query ID.
				// rand.Reader provides randomness for result IDs.
				_, err := inline.New(api, rand.Reader, u.QueryID).
					CacheTimeSeconds(1). // small cache, results depend on query
					Set(ctx,
						// Article result with styled message content.
						inline.Article("Styled greeting",
							inline.MessageStyledText(
								styling.Bold("Hello"),
								styling.Plain(", you searched for "),
								styling.Italic(u.Query),
							),
						).
							ID("styled").
							Description("Sends a styled greeting"),
						// Article result that echoes the query as plain text.
						inline.Article("Echo",
							inline.MessageText(u.Query),
						).
							ID("echo").
							Description("Sends your query back"),
					)
				return err
			})
			return nil
		}, telegram.RunUntilCanceled)
	})
}
