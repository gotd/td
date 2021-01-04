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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
)

func TestConstantAuth(t *testing.T) {
	askCode := telegram.CodeAuthenticatorFunc(func(ctx context.Context) (string, error) {
		return "123", nil
	})

	a := require.New(t)
	auth := telegram.ConstantAuth("phone", "password", askCode)
	ctx := context.Background()

	result, err := auth.Code(ctx)
	a.NoError(err)
	a.Equal("123", result)

	result, err = auth.Phone(ctx)
	a.NoError(err)
	a.Equal("phone", result)

	result, err = auth.Password(ctx)
	a.NoError(err)
	a.Equal("password", result)
}

func TestCodeOnlyAuth(t *testing.T) {
	askCode := telegram.CodeAuthenticatorFunc(func(ctx context.Context) (string, error) {
		return "123", nil
	})

	a := require.New(t)
	auth := telegram.CodeOnlyAuth("phone", askCode)
	ctx := context.Background()

	result, err := auth.Code(ctx)
	a.NoError(err)
	a.Equal("123", result)

	result, err = auth.Phone(ctx)
	a.NoError(err)
	a.Equal("phone", result)

	_, err = auth.Password(ctx)
	a.Error(err)
}

func ExampleTestAuth() {
	// Example of using test server.
	const dcID = 2

	ctx := context.Background()
	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		Addr: telegram.AddrTest,
	})
	go func() {
		if err := client.Run(ctx); err != nil {
			panic(err)
		}
	}()

	if err := telegram.NewAuth(
		telegram.TestAuth(rand.Reader, dcID),
		telegram.SendCodeOptions{},
	).Run(ctx, client); err != nil {
		panic(err)
	}
}

func ExampleAuthFlow_Run() {
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
	go func() { check(client.Run(ctx)) }()

	codeAsk := func(ctx context.Context) (string, error) {
		fmt.Print("code:")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		code = strings.ReplaceAll(code, "\n", "")
		return code, nil
	}

	check(telegram.NewAuth(
		telegram.ConstantAuth(phone, pass, telegram.CodeAuthenticatorFunc(codeAsk)),
		telegram.SendCodeOptions{},
	).Run(ctx, client))
}
