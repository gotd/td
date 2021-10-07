package tdesktop_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nnqq/td/session/tdesktop"
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
		a := account.Authorization
		fmt.Println(a.UserID, a.MainDC)
	}
}
