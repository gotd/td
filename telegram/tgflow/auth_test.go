package tgflow_test

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/tgflow"
)

func TestConstantAuth(t *testing.T) {
	askCode := tgflow.CodeAuthenticatorFunc(func(ctx context.Context) (string, error) {
		return "123", nil
	})

	a := require.New(t)
	auth := tgflow.ConstantAuth("phone", "password", askCode)
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
	askCode := tgflow.CodeAuthenticatorFunc(func(ctx context.Context) (string, error) {
		return "123", nil
	})

	a := require.New(t)
	auth := tgflow.CodeOnlyAuth("phone", askCode)
	ctx := context.Background()

	result, err := auth.Code(ctx)
	a.NoError(err)
	a.Equal("123", result)

	result, err = auth.Phone(ctx)
	a.NoError(err)
	a.Equal("phone", result)

	result, err = auth.Password(ctx)
	a.Error(err)
}

func ExampleAuth_Run() {
	check := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	appIDString := os.Getenv("APP_ID")
	appHash := os.Getenv("APP_HASH")
	phone := os.Getenv("PHONE")
	pass := os.Getenv("PASSWORD")

	if appIDString == "" || appHash == "" || phone == "" || pass == "" {
		log.Fatal("PHONE, PASSWORD, APP_ID or APP_HASH is not set: skip")
	}

	appID, err := strconv.Atoi(appIDString)
	check(err)

	ctx := context.Background()
	client := telegram.NewClient(appID, appHash, telegram.Options{})
	check(client.Connect(ctx))

	codeAsk := func(ctx context.Context) (string, error) {
		fmt.Print("code:")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		code = strings.ReplaceAll(code, "\n", "")
		return code, nil
	}

	check(tgflow.NewAuth(
		tgflow.ConstantAuth(phone, pass, tgflow.CodeAuthenticatorFunc(codeAsk)),
		telegram.SendCodeOptions{},
	).Run(ctx, client))
}
