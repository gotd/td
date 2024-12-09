package telegram_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/td/session"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/internal/e2etest"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

func tryConnect(ctx context.Context, opts telegram.Options) error {
	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, opts)
	return client.Run(ctx, func(ctx context.Context) error {
		_, err := client.API().HelpGetNearestDC(ctx)
		return err
	})
}

func testTransportExternal(resolver dcs.Resolver, storage session.Storage) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		log := zaptest.NewLogger(t)
		defer func() { _ = log.Sync() }()

		require.NoError(t, tryConnect(ctx, telegram.Options{
			Logger:         log.Named("client"),
			SessionStorage: storage,
			Resolver:       resolver,
		}))
	}
}

func TestExternalE2EConnect(t *testing.T) {
	testutil.SkipExternal(t)
	// To re-use session.
	storage := &session.StorageMemory{}

	tcp := func(p dcs.Protocol) func(t *testing.T) {
		return testTransportExternal(dcs.Plain(dcs.PlainOptions{Protocol: p}), storage)
	}

	t.Run("Abridged", tcp(transport.Abridged))
	t.Run("Intermediate", tcp(transport.Intermediate))
	t.Run("PaddedIntermediate", tcp(transport.PaddedIntermediate))
	t.Run("Full", tcp(transport.Full))

	wsOpts := dcs.WebsocketOptions{}
	t.Run("Websocket", testTransportExternal(dcs.Websocket(wsOpts), storage))
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

func TestExternalE2EUsersDialog(t *testing.T) {
	testutil.SkipExternal(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	log := zaptest.NewLogger(t).WithOptions(zap.IncreaseLevel(zapcore.InfoLevel))

	cfg := e2etest.TestOptions{
		Logger: log,
	}
	suite := e2etest.NewSuite(t, cfg)

	auth := make(chan *tg.User, 1)
	g := tdsync.NewLogGroup(ctx, log.Named("group"))

	g.Go("echobot", func(ctx context.Context) error {
		if err := e2etest.NewEchoBot(suite, auth).Run(ctx); err != nil {
			return errors.Wrap(err, "echo bot")
		}
		return nil
	})

	user, ok := <-auth
	if ok {
		g.Go("terentyev", func(ctx context.Context) error {
			defer g.Cancel()
			if err := e2etest.NewUser(suite, strings.Split(dialog, "\n"), user.Username).Run(ctx); err != nil {
				return errors.Wrap(err, "user")
			}
			return nil
		})
	}

	require.NoError(t, g.Wait())
}
