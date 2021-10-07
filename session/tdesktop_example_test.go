package session_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nnqq/td/session"
	"github.com/nnqq/td/session/tdesktop"
	"github.com/nnqq/td/telegram"
)

func ExampleTDesktopSession() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	root := filepath.Join(home, "Downloads", "Telegram", "tdata")
	accounts, err := tdesktop.Read(root, nil)
	if err != nil {
		panic(err)
	}

	data, err := session.TDesktopSession(accounts[0])
	if err != nil {
		panic(err)
	}

	fmt.Println(data.DC, data.Addr)
}

func ExampleTDesktopSession_convert() {
	ctx := context.Background()

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	root := filepath.Join(home, "Downloads", "Telegram", "tdata")
	accounts, err := tdesktop.Read(root, nil)
	if err != nil {
		panic(err)
	}

	data, err := session.TDesktopSession(accounts[0])
	if err != nil {
		panic(err)
	}

	var (
		storage = new(session.StorageMemory)
		loader  = session.Loader{Storage: storage}
	)

	// Save decoded Telegram Desktop session as gotd session.
	if err := loader.Save(ctx, data); err != nil {
		panic(err)
	}

	// Create client.
	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		SessionStorage: storage,
	})
	if err := client.Run(ctx, func(ctx context.Context) error {
		// Use Telegram Desktop session.
		return nil
	}); err != nil {
		panic(err)
	}
}
