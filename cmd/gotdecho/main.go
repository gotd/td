// Binary gotdecho provides example of Telegram echo bot.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/proxy"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

func run(ctx context.Context) error {
	logger, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.DebugLevel))
	defer func() { _ = logger.Sync() }()

	// Reading app id from env (never hardcode it!).
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		return xerrors.Errorf("APP_ID not set or invalid: %w", err)
	}

	appHash := os.Getenv("APP_HASH")
	if appHash == "" {
		return xerrors.New("no APP_HASH provided")
	}

	// Setting up session storage.
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sessionDir := filepath.Join(home, ".td")
	if err := os.MkdirAll(sessionDir, 0600); err != nil {
		return err
	}

	dispatcher := tg.NewUpdateDispatcher()
	client := telegram.NewClient(appID, appHash, telegram.Options{
		Logger: logger,
		SessionStorage: &telegram.FileSessionStorage{
			Path: filepath.Join(sessionDir, "session.json"),
		},

		Transport:     transport.Intermediate(transport.DialFunc(proxy.Dial)),
		UpdateHandler: dispatcher,
	})

	dispatcher.OnNewMessage(func(ctx tg.UpdateContext, u *tg.UpdateNewMessage) error {
		switch m := u.Message.(type) {
		case *tg.Message:
			if m.Out {
				// Skipping updates from self.
				return nil
			}

			switch peer := m.PeerID.(type) {
			case *tg.PeerUser:
				user := ctx.Users[peer.UserID]
				logger.Info("Got message", zap.String("text", m.Message),
					zap.Int("user_id", user.ID),
					zap.String("user_first_name", user.FirstName),
					zap.String("username", user.Username))

				return client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
					Message: m.Message,
					Peer:    user.AsInputPeer(),
				})
			}
		}

		return nil
	})

	return client.Run(ctx, func(ctx context.Context) error {
		self, err := client.Self(ctx)
		if err != nil || !self.Bot {
			if err := client.AuthBot(ctx, os.Getenv("BOT_TOKEN")); err != nil {
				return xerrors.Errorf("failed to perform bot login: %w", err)
			}
			logger.Info("Bot login ok")
		}

		state, err := tg.NewClient(client).UpdatesGetState(ctx)
		if err != nil {
			return xerrors.Errorf("failed to get state: %w", err)
		}
		logger.Sugar().Infof("Got state: %+v", state)

		<-ctx.Done()
		return ctx.Err()
	})
}

func withSignal(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, func() {
		signal.Stop(c)
		cancel()
	}
}

func main() {
	ctx, cancel := withSignal(context.Background())
	defer cancel()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
