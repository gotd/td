// Binary userbot-echo replies to private messages with the same text, as a
// user account, using the telegram/user bootstrap.
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
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/user"
	"github.com/gotd/td/tg"
)

func main() {
	// Environment variables:
	// 	APP_ID:     app_id of Telegram app.
	// 	APP_HASH:   app_hash of Telegram app.
	// 	PHONE:      phone number of the account (prompted if empty); the code
	// 	            and the 2FA password are prompted on the first start only.
	// 	STATE_DIR:  directory for the session and the update state
	// 	            (default "userbot-echo"), so that later starts don't sign in
	// 	            again and get the messages sent while stopped.
	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
		defer stop()

		appID, err := strconv.Atoi(os.Getenv("APP_ID"))
		if err != nil {
			return errors.Wrap(err, "APP_ID")
		}
		dir := os.Getenv("STATE_DIR")
		if dir == "" {
			dir = "userbot-echo"
		}

		d := tg.NewUpdateDispatcher()
		u, err := user.New(appID, os.Getenv("APP_HASH"), user.Options{
			Dir:     dir,
			Handler: d,
			// Used only if the stored session is not authorized.
			Auth:   user.Phone(examples.Terminal{PhoneNumber: os.Getenv("PHONE")}, auth.SendCodeOptions{}),
			Client: telegram.Options{Logger: logzap.New(log)},
		})
		if err != nil {
			return err
		}

		// Private chats: reply with the same text.
		d.OnNewMessage(func(ctx context.Context, e tg.Entities, upd *tg.UpdateNewMessage) error {
			m, ok := upd.Message.(*tg.Message)
			if !ok || m.Out {
				return nil
			}
			if _, ok := m.PeerID.(*tg.PeerUser); !ok {
				return nil
			}
			_, err := u.Sender().Reply(e, upd).Text(ctx, m.Message)
			return err
		})

		// Bootstrap: connects, signs in only if needed, and starts the
		// updates manager.
		if err := u.Start(ctx, telegram.WithLifetime(ctx)); err != nil {
			return err
		}
		defer func() { _ = u.Close() }()
		log.Info("Signed in", zap.Int64("user_id", u.ID()))

		// Runs until Ctrl+C or a fatal error, e.g. a terminated session.
		return u.Wait(context.Background())
	})
}
