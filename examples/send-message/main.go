// Binary send-message logs in as a user with a phone code, persists the
// session to a file, and sends a text message to a channel or user.
//
// This addresses the three most requested scenarios from
// https://github.com/gotd/td/issues/166:
//   - user login by code with session storage in a file;
//   - reconnect using the previously stored session (no code on next run);
//   - sending a message to a channel.
//
// On the first run it prompts for phone, code and 2FA password in the
// terminal and writes the session to --session. On subsequent runs the stored
// session is reused and no login is required.
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strconv"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
)

func run(ctx context.Context) error {
	var (
		appID       = flag.Int("api-id", 0, "app id (or APP_ID env)")
		appHash     = flag.String("api-hash", "", "app hash (or APP_HASH env)")
		sessionFile = flag.String("session", "session.json", "path to session file")
		target      = flag.String("target", "", "channel or user to send to, e.g. @durov")
		text        = flag.String("text", "Hello from gotd!", "message text")
	)
	flag.Parse()
	if *target == "" {
		return errors.New("no --target provided, e.g. --target @durov")
	}

	id, hash := *appID, *appHash
	if id == 0 {
		id, _ = strconv.Atoi(os.Getenv("APP_ID"))
	}
	if hash == "" {
		hash = os.Getenv("APP_HASH")
	}
	if id == 0 || hash == "" {
		return errors.New("set --api-id/--api-hash or APP_ID/APP_HASH")
	}

	log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel))
	defer func() { _ = log.Sync() }()

	// Persist the session to a file so the next run reuses it.
	client := telegram.NewClient(id, hash, telegram.Options{
		Logger:         logzap.New(log),
		SessionStorage: &session.FileStorage{Path: *sessionFile},
	})

	// Terminal authenticator prompts for phone, code and 2FA password.
	flow := auth.NewFlow(examples.Terminal{}, auth.SendCodeOptions{})

	return client.Run(ctx, func(ctx context.Context) error {
		// Status tells us whether the stored session is still authorized.
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return errors.Wrap(err, "auth status")
		}
		if status.Authorized {
			log.Info("Reusing stored session, no login required")
		} else {
			log.Info("No valid session, logging in with code")
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return errors.Wrap(err, "auth")
			}
		}

		// Resolve the target (@username of a user or channel) and send text.
		sender := message.NewSender(client.API())
		if _, err := sender.Resolve(*target).Text(ctx, *text); err != nil {
			return errors.Wrap(err, "send")
		}
		log.Info("Message sent", zap.String("target", *target))
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
