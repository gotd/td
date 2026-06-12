// Binary takeout exports account data using the telegram/takeout helper.
//
// Telegram requires a dedicated "takeout" session for bulk data export (the
// same mechanism the official "Export Telegram data" feature uses). The
// takeout.Run helper initializes such a session, runs your callback with a
// wrapped invoker, and finalizes the session afterwards.
//
// This example exports the contact list and prints it.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/takeout"
	"github.com/gotd/td/tg"
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

		// Configure what to export. Here we only request contacts.
		cfg := takeout.Config{Contacts: true}

		// takeout.Run wraps all API calls made inside the callback with the
		// takeout session and finishes it when the callback returns.
		return takeout.Run(ctx, client, cfg, func(ctx context.Context, t *takeout.Client) error {
			log.Info("Takeout session started", zap.Int64("id", t.ID()))

			// Use tg.NewClient over the takeout client to call API methods.
			api := tg.NewClient(t)

			result, err := api.ContactsGetContacts(ctx, 0)
			if err != nil {
				return errors.Wrap(err, "get contacts")
			}

			contacts, ok := result.(*tg.ContactsContacts)
			if !ok {
				log.Info("Contacts not modified")
				return nil
			}

			log.Info("Got contacts", zap.Int("count", len(contacts.Users)))
			for _, u := range contacts.Users {
				user, ok := u.AsNotEmpty()
				if !ok {
					continue
				}
				username, _ := user.GetUsername()
				phone, _ := user.GetPhone()
				fmt.Printf("%d\t%s\t%s %s\t%s\n", user.ID, username, user.FirstName, user.LastName, phone)
			}
			return nil
		})
	})
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}
