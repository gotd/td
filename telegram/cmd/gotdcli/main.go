package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/xerrors"

	"github.com/ernado/td/telegram"
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

	// Trying to log in as bot.
	if err := client.BotLogin(ctx, telegram.BotLogin{
		ID:    appID,
		Hash:  appHash,
		Token: os.Getenv("BOT_TOKEN"),
	}); err != nil {
		return xerrors.Errorf("failed to perform bot login: %w", err)
	}

	return nil
}

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
