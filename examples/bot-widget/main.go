// Binary bot-widget implements a bot that replies to incoming messages in
// private chats and basic groups with a counter HTML widget and answers
// queries sent by the widget.
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
// bubble and expose the TelegramWidget object to its script (see the
// widget-sandbox docs vendored by the official clients):
//
//   - TelegramWidget.publicData is the public data set by the bot;
//   - TelegramWidget.request(data) sends a query to the bot (it arrives as
//     updateBotWidgetQuery with Data set to data) and resolves to {result};
//   - the "publicDataChanged" event fires when the bot edits public data.
//
// The widget sends the current count; the bot answers with the new one and
// stores it in public data, so every viewer of the message sees it.
const widgetHTML = `<!DOCTYPE html>
<html><head><meta charset="utf-8"></head>
<body>
<button id="b">0</button>
<script>
var T = TelegramWidget, b = document.getElementById('b');
function count() { try { return JSON.parse(T.publicData || '{}').n || 0; } catch (e) { return 0; } }
function render() { b.textContent = 'Clicked ' + count() + ' times'; }
b.addEventListener('click', function () {
  T.request(String(count())).then(render, render);
});
T.onEvent('publicDataChanged', render);
render();
</script>
</body></html>`

func main() {
	// Environment variables:
	//	BOT_TOKEN:     token from BotFather
	// 	APP_ID:        app_id of Telegram app.
	// 	APP_HASH:      app_hash of Telegram app.
	// 	SESSION_FILE:  path to session file
	// 	SESSION_DIR:   path to session directory, if SESSION_FILE is not set
	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		// Fail fast on HTML that exceeds client-side widget limits. The
		// compressed size limit is enforced by the server; ValidateCompressed
		// approximates it.
		if err := widget.ValidateHTML(widgetHTML); err != nil {
			return err
		}
		if err := widget.ValidateCompressed(widgetHTML); err != nil {
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

			// Reply to messages in private chats and basic groups with a
			// widget (supergroups would need OnNewChannelMessage).
			dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
				m, ok := u.Message.(*tg.Message)
				if !ok || m.Out {
					return nil
				}

				_, err := sender.Reply(e, u).Media(ctx,
					message.Widget(widgetHTML, styling.Plain("Counter widget")).
						PublicData(`{"n":0}`). // visible in forwards and copies
						AllowQueries(true),    // let the widget query the bot
				)
				return err
			})

			// The widget sends its current count as query data: answer with
			// the incremented count and store it in the widget public data.
			dispatcher.OnBotWidgetQuery(func(ctx context.Context, e tg.Entities, u *tg.UpdateBotWidgetQuery) error {
				log.Info("Widget query",
					zap.Int64("user_id", u.UserID),
					zap.Int("msg_id", u.MsgID),
					zap.ByteString("data", u.Data),
				)

				n, err := strconv.Atoi(string(u.Data))
				if err != nil {
					// Unknown query: answer without data.
					return sender.AnswerWidgetQuery(ctx, u.QueryID, "")
				}
				state := `{"n":` + strconv.Itoa(n+1) + `}`

				// The update carries the chat peer, so Peer resolves it from
				// update entities.
				if _, err := sender.Peer(e, u).Edit(u.MsgID).WidgetData(ctx,
					message.WidgetPublicData(state),
				); err != nil {
					return err
				}
				return sender.AnswerWidgetQuery(ctx, u.QueryID, state)
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
