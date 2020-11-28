package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/xerrors"

	"github.com/ernado/td/crypto"
	"github.com/ernado/td/telegram"
	"github.com/ernado/td/tg"
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

	// Creating connection.
	client, err := telegram.Dial(ctx, telegram.Options{
		Addr:   "149.154.167.50:443",
		Logger: logger,
	})
	if err != nil {
		return xerrors.Errorf("failed to dial: %w", err)
	}

	// Connecting. This will execute key exchange and start read loop.
	if err := client.Connect(ctx); err != nil {
		return xerrors.Errorf("failed to connect: %w", err)
	}
	logger.Info("ok")

	// Initialize connection via initConnection rpc call.
	if err := client.InitConnection(ctx, telegram.Init{
		AppID: appID,
	}); err != nil {
		return xerrors.Errorf("failed to init connection: %w", err)
	}

	client.SetUpdateHandler(func(updates *tg.Updates) error {
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
						logger.With(
							zap.String("text", m.Message),
							zap.Int("user_id", user.ID),
							zap.String("user_first_name", user.FirstName),
							zap.String("username", user.Username),
						).Info("Got message")

						randomID, err := crypto.RandInt64(rand.Reader)
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
				logger.With(zap.String("update_type", fmt.Sprintf("%T", u))).Info("Ignoring update")
			}
		}
		return nil
	})

	// Trying to log in as bot.
	if err := client.BotLogin(ctx, telegram.BotLogin{
		ID:    appID,
		Hash:  appHash,
		Token: os.Getenv("BOT_TOKEN"),
	}); err != nil {
		return xerrors.Errorf("failed to perform bot login: %w", err)
	}

	// Just reading updates.
	select {}
}

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
