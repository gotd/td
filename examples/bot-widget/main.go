// Binary bot-widget implements a bot that replies to any private message with
// an HTML widget and answers queries sent by the widget.
//
// It demonstrates message.Widget (sending a widget with public data and
// queries allowed), Sender.AnswerWidgetQuery (answering widget queries) and
// EditMessageBuilder.WidgetData (updating data of a sent widget).
package main

import (
	"context"
	"strconv"

	"go.uber.org/zap"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/message/widget"
	"github.com/gotd/td/tg"
)

// widgetHTML is the widget source. Clients render it inside the message
// bubble; public data and chat data are available to the widget script, which
// can also send queries to the bot.
const widgetHTML = `<!doctype html>
<html>
<body style="margin:0;font-family:sans-serif">
  <p>Hello from a widget!</p>
</body>
</html>`

func main() {
	// Environment variables:
	//	BOT_TOKEN:     token from BotFather
	// 	APP_ID:        app_id of Telegram app.
	// 	APP_HASH:      app_hash of Telegram app.
	// 	SESSION_FILE:  path to session file
	// 	SESSION_DIR:   path to session directory, if SESSION_FILE is not set
	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		// Fail fast on HTML that exceeds widget limits.
		if err := widget.ValidateHTML(widgetHTML); err != nil {
			return err
		}

		dispatcher := tg.NewUpdateDispatcher()
		opts := telegram.Options{
			Logger:        logzap.New(log),
			UpdateHandler: dispatcher,
		}
		return telegram.BotFromEnvironment(ctx, opts, func(ctx context.Context, client *telegram.Client) error {
			api := tg.NewClient(client)
			sender := message.NewSender(api)

			// Reply to private messages with a widget.
			dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
				m, ok := u.Message.(*tg.Message)
				if !ok || m.Out {
					return nil
				}

				_, err := sender.Reply(e, u).Media(ctx,
					message.Widget(widgetHTML, styling.Plain("Counter widget")).
						Size(0, 120).          // full bubble width, 120px height
						PublicData(`{"n":0}`). // visible in forwards and copies
						AllowQueries(true),    // let the widget query the bot
				)
				return err
			})

			// Answer widget queries: increment the counter and update widget
			// public data, so every viewer of the message sees the new value.
			dispatcher.OnBotWidgetQuery(func(ctx context.Context, e tg.Entities, u *tg.UpdateBotWidgetQuery) error {
				log.Info("Widget query",
					zap.Int64("user_id", u.UserID),
					zap.Int("msg_id", u.MsgID),
					zap.ByteString("data", u.Data),
				)

				n, _ := strconv.Atoi(string(u.Data))
				n++
				if err := sender.AnswerWidgetQuery(ctx, u.QueryID, strconv.Itoa(n)); err != nil {
					return err
				}

				// The update carries the chat peer, so Peer resolves it from
				// update entities.
				_, err := sender.Peer(e, u).Edit(u.MsgID).WidgetData(ctx,
					message.WidgetPublicData(`{"n":`+strconv.Itoa(n)+`}`),
				)
				return err
			})

			// Edited widget data is delivered as updateMessageWidgetData.
			dispatcher.OnMessageWidgetData(func(ctx context.Context, e tg.Entities, u *tg.UpdateMessageWidgetData) error {
				log.Info("Widget data updated", zap.Int("msg_id", u.MsgID), zap.String("public_data", u.PublicData))
				return nil
			})
			return nil
		}, telegram.RunUntilCanceled)
	})
}
