package telegram_test

import (
	"context"
	"crypto/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/mtproto/tgtest"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/internal/e2etest"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

func testTransport(trp mtproto.Transport) func(t *testing.T) {
	return func(t *testing.T) {
		log := zaptest.NewLogger(t)
		defer func() { _ = log.Sync() }()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		client, err := telegram.New(mtproto.TestAppID, mtproto.TestAppHash, telegram.Options{
			MTProto: mtproto.Options{
				Addr:      mtproto.AddrTest,
				Transport: trp,
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = client.Close()
		}()

		if err := telegram.NewAuth(
			telegram.TestAuth(rand.Reader, 2),
			telegram.SendCodeOptions{},
		).Run(ctx, client); err != nil {
			t.Fatal(err)
		}

		if _, err := client.Self(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

func TestExternalE2EConnect(t *testing.T) {
	if ok, _ := strconv.ParseBool(os.Getenv("GOTD_TEST_EXTERNAL")); !ok {
		t.Skip("Skipped. Set GOTD_TEST_EXTERNAL=1 to enable external e2e test.")
	}

	t.Run("abridged", testTransport(transport.Abridged(nil)))
	t.Run("intermediate", testTransport(transport.Intermediate(nil)))
	t.Run("padded intermediate", testTransport(transport.PaddedIntermediate(nil)))
	t.Run("full", testTransport(transport.Full(nil)))
}

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

func TestE2EUsersDialog(t *testing.T) {
	if ok, _ := strconv.ParseBool(os.Getenv("GOTD_USERS_DIALOG")); !ok {
		t.Skip("Skipped. Set GOTD_USERS_DIALOG=1 to enable users dialog e2e test.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	log := zaptest.NewLogger(t).WithOptions(zap.IncreaseLevel(zapcore.InfoLevel))

	cfg := e2etest.TestConfig{
		AppID:   mtproto.TestAppID,
		AppHash: mtproto.TestAppHash,
		DcID:    2,
		Addr:    mtproto.AddrTest,
	}
	suite := e2etest.NewSuite(tgtest.NewSuite(ctx, t, log), cfg, rand.Reader)

	g, gctx := errgroup.WithContext(ctx)
	auth := make(chan *tg.User, 1)
	botCtx, botCancel := context.WithCancel(gctx)
	g.Go(func() error {
		return e2etest.NewEchoBot(suite, auth).Run(botCtx)
	})

	user := <-auth
	g.Go(func() error {
		defer botCancel()
		return e2etest.NewUser(suite, strings.Split(dialog, "\n"), user).Run(gctx)
	})

	require.NoError(t, g.Wait())
}
