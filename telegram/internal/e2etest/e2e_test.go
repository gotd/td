package e2etest

import (
	"context"
	"log"
	"strings"
	"testing"
)

const dialog = `— Да?
— Алё!
— Да да?
— Ну как там с деньгами?
— А?
— Как с деньгами-то там?
— Чё с деньгами?
— Чё?
— Куда ты звонишь?
— Тебе звоню.
— Кому?
— Ну тебе.`

func TestWithTestDC(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	suite := NewSuite(
		t,
		// from https://github.com/telegramdesktop/tdesktop/blob/master/docs/api_credentials.md
		17349,
		"344583e45741c457fe1862106095a5eb",
		// from https://docs.telethon.dev/en/latest/developing/test-servers.html
		2,
		"149.154.167.40:80",
	)

	ctx := context.Background()
	creator := NewBotCreator(suite)
	err := creator.Connect(ctx)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	name := "gotd_super_bot"
	token, err := creator.CreateBot(ctx, name)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	t.Cleanup(func() {
		_ = creator.DeleteBot(ctx, name)
	})

	bot := NewEchoBot(suite, token)
	go func() {
		err := bot.Run(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()
	defer bot.Stop()

	father, err := creator.ResolveUsername(ctx, name)
	if err != nil {
		t.Fatal(err)
	}

	user := NewUser(suite, strings.Split(dialog, "\n"), father)
	err = user.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
