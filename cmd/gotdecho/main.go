package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type updateHandler struct {
	log *zap.Logger
}

func (h updateHandler) handle(ctx context.Context, client telegram.UpdateClient, updates *tg.Updates) error {
	// This wll be required to send message back.
	users := map[int]*tg.User{}
	for _, u := range updates.Users {
		user, ok := u.(*tg.User)
		if !ok {
			continue
		}
		users[user.ID] = user
	}

	for _, update := range updates.Updates {
		switch u := update.(type) {
		case *tg.UpdateNewMessage:
			switch m := u.Message.(type) {
			case *tg.Message:
				switch peer := m.PeerID.(type) {
				case *tg.PeerUser:
					user := users[peer.UserID]
					h.log.With(
						zap.String("text", m.Message),
						zap.Int("user_id", user.ID),
						zap.String("user_first_name", user.FirstName),
						zap.String("username", user.Username),
					).Info("Got message")

					randomID, err := client.RandInt64()
					if err != nil {
						return err
					}
					return client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
						RandomID: randomID,
						Message:  m.Message,
						Peer: &tg.InputPeerUser{
							UserID:     user.ID,
							AccessHash: user.AccessHash,
						},
					})
				}
			}
		default:
			h.log.With(zap.String("update_type", fmt.Sprintf("%T", u))).Info("Ignoring update")
		}
	}
	return nil
}

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
	sessionDir := path.Join(home, ".td")
	if err := os.MkdirAll(sessionDir, 0600); err != nil {
		return err
	}

	// Creating connection.
	dialCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	client, err := telegram.Dial(dialCtx, telegram.Options{
		Addr:   "149.154.167.50:443",
		Logger: logger,
		SessionStorage: &telegram.FileSessionStorage{
			Path: path.Join(sessionDir, "session.json"),
		},

		// Grab these from https://my.telegram.org/apps.
		// Never share it or hardcode!
		AppID:   appID,
		AppHash: appHash,

		UpdateHandler: updateHandler{log: logger}.handle,
	})
	if err != nil {
		return xerrors.Errorf("failed to dial: %w", err)
	}
	logger.Info("Dialed")

	auth, err := client.AuthStatus(dialCtx)
	if err != nil {
		return xerrors.Errorf("failed to get auth status: %w", err)
	}
	logger.With(zap.Bool("authorized", auth.Authorized)).Info("Auth status")
	if !auth.Authorized {
		if err := client.BotLogin(dialCtx, os.Getenv("BOT_TOKEN")); err != nil {
			return xerrors.Errorf("failed to perform bot login: %w", err)
		}
		logger.Info("Bot login ok")
	}

	// Using tg.Client for directly calling RPC.
	raw := tg.NewClient(client)

	// Getting state is required to process updates in your code.
	// Currently missed updates are not processed, so only new
	// messages will be handled.
	state, err := raw.UpdatesGetState(ctx, &tg.UpdatesGetStateRequest{})
	if err != nil {
		return xerrors.Errorf("failed to get state: %w", err)
	}
	logger.Sugar().Infof("Got state: %+v", state)

	// Reading updates until SIGTERM.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	logger.Info("Shutting down")
	if err := client.Close(ctx); err != nil {
		return err
	}
	logger.Info("Graceful shutdown completed")
	return nil
}

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
