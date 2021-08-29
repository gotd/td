package main

import (
	"context"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel), zap.AddStacktrace(zapcore.FatalLevel))
	defer func() { _ = log.Sync() }()

	gaps := updates.New(updates.Config{
		Handler: func(u tg.UpdatesClass) error {
			log.Info("Updates", zap.Any("updates", u))
			return nil
		},
		Logger: log.Named("gaps"),
	})

	// Initializing client from environment.
	// Available environment variables:
	// 	APP_ID:         app_id of Telegram app.
	// 	APP_HASH:       app_hash of Telegram app.
	// 	SESSION_FILE:   path to session file
	// 	SESSION_DIR:    path to session directory, if SESSION_FILE is not set
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: log,
		UpdateHandler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
			return gaps.HandleUpdates(u)
		}),
		Middlewares: []telegram.Middleware{
			updhook.UpdateHook(telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
				return gaps.HandleUpdates(u)
			})),
		},
	})
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		// Note: you need to be authenticated here.

		// Fetch user info.
		user, err := client.Self(ctx)
		if err != nil {
			return err
		}

		// Notify update manager about authentication.
		if err := gaps.Auth(client.API(), user.ID, user.Bot, true); err != nil {
			return err
		}

		<-ctx.Done()
		return ctx.Err()
	})
}
