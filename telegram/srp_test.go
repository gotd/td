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

func ExampleClient_PasswordLogin() {
	die := func(err error) {
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
	die(err)

	ctxt := context.Background()
	c := telegram.NewClient(appID, appHash, telegram.Options{})
	err = c.Connect(ctxt)
	die(err)

	client := tg.NewClient(c)
	sentCode, err := client.AuthSendCode(ctxt, &tg.AuthSendCodeRequest{
		PhoneNumber: phone,
		APIID:       appID,
		APIHash:     appHash,
		Settings:    tg.CodeSettings{},
	})
	die(err)

	fmt.Print("code:")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	die(err)
	code = strings.ReplaceAll(code, "\n", "")

	_, err = client.AuthSignIn(ctxt, &tg.AuthSignInRequest{
		PhoneNumber:   phone,
		PhoneCodeHash: sentCode.PhoneCodeHash,
		PhoneCode:     code,
	})
	if err != nil && !strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") {
		die(err)
	}

	_, err = c.PasswordLogin(ctxt, pass)
	die(err)
}
