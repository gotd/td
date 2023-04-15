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

	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
		Logger:  log.Named("gaps"),
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

	// Setup message update handlers.
	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		log.Info("Channel message", zap.Any("message", update.Message))
		return nil
	})
	d.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		log.Info("Message", zap.Any("message", update.Message))
		return nil
	})

	return client.Run(ctx, func(ctx context.Context) error {
		// Note: you need to be authenticated here.

		// Fetch user info.
		user, err := client.Self(ctx)
		if err != nil {
			return err
		}
		return gaps.Run(ctx, client.API(), user.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				log.Info("Gaps started")
			},
		})
	})
}
