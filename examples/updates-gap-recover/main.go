package main

import (
	"context"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/updates"
	updhook "github.com/nnqq/td/telegram/updates/hook"
	"github.com/nnqq/td/tg"
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
		Handler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
			log.Info("Updates", zap.Any("updates", u))
			return nil
		}),
		Logger: log.Named("gaps"),
	})

	// Initializing client from environment.
	// Available environment variables:
	// 	APP_ID:         app_id of Telegram app.
	// 	APP_HASH:       app_hash of Telegram app.
	// 	SESSION_FILE:   path to session file
	// 	SESSION_DIR:    path to session directory, if SESSION_FILE is not set
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger:        log,
		UpdateHandler: gaps,
		Middlewares: []telegram.Middleware{
			updhook.UpdateHook(gaps.Handle),
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
		if err := gaps.Auth(ctx, client.API(), user.ID, user.Bot, true); err != nil {
			return err
		}
		defer func() { _ = gaps.Logout() }()

		<-ctx.Done()
		return ctx.Err()
	})
}
