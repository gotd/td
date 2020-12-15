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
	"github.com/gotd/td/tg"
)

func ExampleClient_UserLogin() {
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

	ctxt := context.Background()
	c := telegram.NewClient(appID, appHash, telegram.Options{})
	err = c.Connect(ctxt)
	check(err)

	codeAsk := func(ctx context.Context) (string, error) {
		fmt.Print("code:")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		code = strings.ReplaceAll(code, "\n", "")
		return code, nil
	}
	auth := telegram.ConstantAuth(phone, pass, telegram.CodeAuthenticatorFunc(codeAsk))

	err = c.UserLogin(ctxt, auth, tg.CodeSettings{})
	check(err)
}
