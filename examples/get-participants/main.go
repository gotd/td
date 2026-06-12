// Binary get-participants lists all participants of a channel or supergroup
// using the query/channels/participants iterator helper.
//
// It demonstrates two helpers:
//   - peers.Manager, to resolve a @username into a channel;
//   - query/channels/participants, to page through channels.getParticipants
//     transparently (the helper handles offsets and access hashes for you).
package main

import (
	"context"
	"flag"
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
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query/channels/participants"
)

func run(ctx context.Context) error {
	target := flag.String("target", "", "channel or supergroup @username to list participants of")
	flag.Parse()
	if *target == "" {
		return errors.New("no --target provided, e.g. --target gotd_ru")
	}

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

	api := client.API()
	// Peer manager helper, resolves usernames and caches access hashes.
	manager := peers.Options{Logger: logzap.New(log)}.Build(api)

	return client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return errors.Wrap(err, "auth")
		}

		// Resolve @username to a peer and ensure it is a channel.
		p, err := manager.Resolve(ctx, *target)
		if err != nil {
			return errors.Wrap(err, "resolve")
		}
		channel, ok := p.(peers.Channel)
		if !ok {
			return errors.Errorf("%q is not a channel", *target)
		}

		// Build the participants query with the "recent" filter and iterate.
		// Other filters are available: Admins(), Bots(), Search(q), etc.
		query := participants.NewQueryBuilder(api).
			GetParticipants(channel.InputChannel()).
			Recent()

		if total, err := query.Count(ctx); err == nil {
			log.Info("Total participants", zap.Int("count", total))
		}

		// ForEach pages through all participants for us.
		return query.ForEach(ctx, func(ctx context.Context, elem participants.Elem) error {
			user, ok := elem.User()
			if !ok {
				return nil
			}
			username, _ := user.GetUsername()
			fmt.Printf("%d\t%s\t%s %s\n", user.ID, username, user.FirstName, user.LastName)
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
