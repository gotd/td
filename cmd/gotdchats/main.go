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

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
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
	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		Addr:           telegram.AddrTest,
		Logger:         logger,
		SessionStorage: storage,
		Transport:      transport.Intermediate(transport.DialFunc(proxy.Dial)),
		UpdateHandler:  dispatcher,
	})

	return client.Run(ctx, func(ctx context.Context) error {
		if self, err := client.Self(ctx); err != nil || self.Bot {
			if err := telegram.NewAuth(
				telegram.TestAuth(rand.Reader, *dcID), telegram.SendCodeOptions{},
			).Run(ctx, client); err != nil {
				return xerrors.Errorf("auth: %w", err)
			}
		}

		c := tg.NewClient(client)
		for range time.NewTicker(time.Second * 5).C {
			chats, err := c.MessagesGetAllChats(ctx, nil)

			var rpcErr *mtproto.Error
			if errors.As(err, &rpcErr) && rpcErr.Type == "FLOOD_WAIT" {
				// Server told us to wait N seconds before sending next message.
				logger.Info("Sleeping", zap.Int("seconds", rpcErr.Argument))
				time.Sleep(time.Second * time.Duration(rpcErr.Argument))
				continue
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
	})
}

func main() {
	if err := run(context.Background()); err != nil {
		panic(err)
	}
}
