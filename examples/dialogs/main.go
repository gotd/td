// Binary dialogs lists your dialogs and optionally dumps recent messages of
// Saved Messages, using the telegram/query iterator helpers.
//
// It demonstrates:
//   - query.GetDialogs, an iterator over messages.getDialogs that pages
//     through every dialog and exposes resolved peer entities;
//   - query.Messages(...).GetHistory, an iterator over messages.getHistory.
//
// Both helpers handle offsets, batching and pagination hashes for you.
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
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
)

// errDone stops an iterator early.
var errDone = errors.New("done")

// dialogName resolves a human-readable name for a dialog from its entities.
func dialogName(elem dialogs.Elem) string {
	peer, ok := elem.Dialog.(*tg.Dialog)
	if !ok {
		return "?"
	}
	switch p := peer.Peer.(type) {
	case *tg.PeerUser:
		if u, ok := elem.Entities.User(p.UserID); ok {
			return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
		}
	case *tg.PeerChat:
		if c, ok := elem.Entities.Chat(p.ChatID); ok {
			return c.Title
		}
	case *tg.PeerChannel:
		if c, ok := elem.Entities.Channel(p.ChannelID); ok {
			return c.Title
		}
	}
	return "?"
}

func run(ctx context.Context) error {
	history := flag.Int("history", 0, "if > 0, dump that many recent messages from Saved Messages")
	flag.Parse()

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
		api := client.API()

		// Iterate over every dialog, printing its name.
		count := 0
		if err := query.GetDialogs(api).ForEach(ctx, func(ctx context.Context, elem dialogs.Elem) error {
			if elem.Deleted() {
				return nil
			}
			count++
			fmt.Printf("%3d  %s\n", count, dialogName(elem))
			return nil
		}); err != nil {
			return errors.Wrap(err, "dialogs")
		}
		log.Info("Listed dialogs", zap.Int("count", count))

		// Optionally dump recent messages from Saved Messages.
		if *history > 0 {
			fmt.Println("\nSaved Messages:")
			seen := 0
			err := query.Messages(api).GetHistory(&tg.InputPeerSelf{}).
				ForEach(ctx, func(ctx context.Context, elem messages.Elem) error {
					if seen >= *history {
						// Stop the iterator early.
						return errDone
					}
					if msg, ok := elem.Msg.(*tg.Message); ok {
						seen++
						fmt.Printf("  [%d] %s\n", msg.ID, msg.Message)
					}
					return nil
				})
			if err != nil && !errors.Is(err, errDone) {
				return errors.Wrap(err, "history")
			}
		}
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
