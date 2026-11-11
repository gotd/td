// Binary bot-echo implements basic example for bot.
package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/bot"
	"github.com/gotd/td/tg"
)

func main() {
	// Environment variables:
	//	BOT_TOKEN:  token from BotFather
	// 	APP_ID:     app_id of Telegram app.
	// 	APP_HASH:   app_hash of Telegram app.
	// 	STATE_DIR:  directory for the session and the update state
	//	            (default "bot-echo"), so that the bot gets the messages
	//	            sent while it was stopped.
	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
		defer stop()

		appID, err := strconv.Atoi(os.Getenv("APP_ID"))
		if err != nil {
			return errors.Wrap(err, "APP_ID")
		}
		dir := os.Getenv("STATE_DIR")
		if dir == "" {
			dir = "bot-echo"
		}

		// Dispatcher handles incoming updates.
		d := tg.NewUpdateDispatcher()
		b, err := bot.New(appID, os.Getenv("APP_HASH"), os.Getenv("BOT_TOKEN"), bot.Options{
			Dir:     dir,
			Handler: d,
			Client:  telegram.Options{Logger: logzap.New(log)},
		})
		if err != nil {
			return err
		}

		// Private chats and basic groups.
		d.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
			m, ok := u.Message.(*tg.Message)
			if !ok || m.Out {
				// Outgoing message, not interesting.
				return nil
			}
			_, err := b.Sender().Reply(e, u).Text(ctx, m.Message)
			return err
		})
		// Supergroups and channels.
		d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewChannelMessage) error {
			m, ok := u.Message.(*tg.Message)
			if !ok || m.Out {
				return nil
			}
			_, err := b.Sender().Reply(e, u).Text(ctx, m.Message)
			return err
		})

		// Bootstrap: connects, signs in (or checks that the stored session
		// is of this bot) and starts the updates manager.
		if err := b.Start(ctx, telegram.WithLifetime(ctx)); err != nil {
			return err
		}
		defer func() { _ = b.Close() }()
		log.Info("Bot started", zap.String("username", b.Self().Username))

		// Runs until Ctrl+C or a fatal error, e.g. a revoked token.
		return b.Wait(context.Background())
	})
}
