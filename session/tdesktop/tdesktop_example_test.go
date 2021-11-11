package tdesktop_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotd/td/session/tdesktop"
)

func ExampleRead() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	root := filepath.Join(home, "Downloads", "Telegram", "tdata")
	accounts, err := tdesktop.Read(root, nil)
	if err != nil {
		panic(err)
	}

	for _, account := range accounts {
		auth := account.Authorization
		cfg := account.Config
		fmt.Println(auth.UserID, auth.MainDC, cfg.Environment)
	}
}
