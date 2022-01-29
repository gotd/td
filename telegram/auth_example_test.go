package telegram_test

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

func ExampleClient_Auth_codeOnly() {
	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	var (
		appIDString = os.Getenv("APP_ID")
		appHash     = os.Getenv("APP_HASH")
		phone       = os.Getenv("PHONE")
	)
	if appIDString == "" || appHash == "" || phone == "" {
		log.Fatal("PHONE, APP_ID or APP_HASH is not set")
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
			auth.CodeOnly(phone, auth.CodeAuthenticatorFunc(codeAsk)),
			auth.SendCodeOptions{},
		).Run(ctx, client.Auth())
	}))
}

func ExampleClient_Auth_password() {
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
		return client.Auth().Test(ctx, dcID)
	}); err != nil {
		panic(err)
	}
}

func ExampleClient_Auth_bot() {
	ctx := context.Background()
	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{})
	if err := client.Run(ctx, func(ctx context.Context) error {
		// Checking auth status.
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return err
		}
		// Can be already authenticated if we have valid session in
		// session storage.
		if !status.Authorized {
			// Otherwise, perform bot authentication.
			if _, err := client.Auth().Bot(ctx, os.Getenv("BOT_TOKEN")); err != nil {
				return err
			}
		}

		// All good, manually authenticated.
		return nil
	}); err != nil {
		panic(err)
	}
}

func ExampleQR_Auth() {
	ctx := context.Background()

	d := tg.NewUpdateDispatcher()
	loggedIn := qrlogin.OnLoginToken(d)
	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		UpdateHandler: d,
	})
	if err := client.Run(ctx, func(ctx context.Context) error {
		qr := client.QR()
		authorization, err := qr.Auth(ctx, loggedIn, func(ctx context.Context, token qrlogin.Token) error {
			fmt.Printf("Open %s using your phone\n", token.URL())
			return nil
		})
		if err != nil {
			return err
		}

		u, ok := authorization.User.AsNotEmpty()
		if !ok {
			return fmt.Errorf("unexpected type %T", authorization.User)
		}
		fmt.Println("ID:", u.ID, "Username:", u.Username, "Bot:", u.Bot)
		return nil
	}); err != nil {
		panic(err)
	}
}
