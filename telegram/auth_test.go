package telegram_test

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/auth"
	"github.com/nnqq/td/telegram/dcs"
	"github.com/nnqq/td/tg"
)

func ExampleClient_Auth() {
	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	var (
		appIDString = os.Getenv("APP_ID")
		appHash     = os.Getenv("APP_HASH")
		phone       = os.Getenv("PHONE")
		pass        = os.Getenv("PASSWORD")
	)
	if appIDString == "" || appHash == "" || phone == "" || pass == "" {
		log.Fatal("PHONE, PASSWORD, APP_ID or APP_HASH is not set")
	}

	appID, err := strconv.Atoi(appIDString)
	check(err)

	ctx := context.Background()
	client := telegram.NewClient(appID, appHash, telegram.Options{})
	codeAsk := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		fmt.Print("code:")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		code = strings.ReplaceAll(code, "\n", "")
		return code, nil
	}

	check(client.Run(ctx, func(ctx context.Context) error {
		return auth.NewFlow(
			auth.Constant(phone, pass, auth.CodeAuthenticatorFunc(codeAsk)),
			auth.SendCodeOptions{},
		).Run(ctx, client.Auth())
	}))
}

func ExampleClient_Auth_test() {
	// Example of using test server.
	const dcID = 2

	ctx := context.Background()
	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		DC:     dcID,
		DCList: dcs.Test(),
	})
	if err := client.Run(ctx, func(ctx context.Context) error {
		return auth.NewFlow(
			auth.Test(rand.Reader, dcID),
			auth.SendCodeOptions{},
		).Run(ctx, client.Auth())
	}); err != nil {
		panic(err)
	}
}
