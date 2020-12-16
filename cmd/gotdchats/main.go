// Binary gotdchats implements chat list request example using testing server.
package main

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/proxy"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/tgflow"
	"github.com/gotd/td/tg"
)

func run(ctx context.Context) error {
	logger, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.DebugLevel))
	defer func() { _ = logger.Sync() }()

	var (
		persisted = flag.Bool("persisted", false, "persist session")
		dcID      = flag.Int("dc", 2, "ID of DC")
	)
	flag.Parse()

	var storage telegram.SessionStorage
	if *persisted {
		// Setting up session storage.
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		sessionDir := filepath.Join(home, ".td")
		if err := os.MkdirAll(sessionDir, 0600); err != nil {
			return err
		}
		storage = &telegram.FileSessionStorage{
			Path: filepath.Join(sessionDir, "session-user.json"),
		}
	}

	dispatcher := tg.NewUpdateDispatcher()
	// Creating connection.
	dialCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		Addr:           telegram.AddrTest,
		Logger:         logger,
		SessionStorage: storage,
		Dialer:         telegram.DialFunc(proxy.Dial),
		UpdateHandler:  dispatcher.Handle,
	})

	if err := client.Connect(dialCtx); err != nil {
		return xerrors.Errorf("failed to connect: %w", err)
	}

	if self, err := client.Self(ctx); err != nil || self.Bot {
		if err := tgflow.NewAuth(tgflow.TestAuth(rand.Reader, *dcID), telegram.SendCodeOptions{}).Run(ctx, client); err != nil {
			return xerrors.Errorf("failed to auth: %w", err)
		}
	}

	c := tg.NewClient(client)

	for range time.NewTicker(time.Second * 5).C {
		chats, err := c.MessagesGetAllChats(ctx, nil)

		var rpcErr *telegram.Error
		if errors.As(err, &rpcErr) && rpcErr.Type == "FLOOD_WAIT" {
			// Server told us to wait N seconds before sending next message.
			logger.With(zap.Int("seconds", rpcErr.Argument)).Info("Sleeping")
			time.Sleep(time.Second * time.Duration(rpcErr.Argument))
		}

		if err != nil {
			return xerrors.Errorf("failed to get chats: %w", err)
		}

		switch chats.(type) {
		case *tg.MessagesChats: // messages.chats#64ff9fd5
			logger.Info("Chats")
		case *tg.MessagesChatsSlice: // messages.chatsSlice#9cd81144
			logger.Info("Slice")
		}
	}

	return nil
}

func main() {
	if err := run(context.Background()); err != nil {
		panic(err)
	}
}
