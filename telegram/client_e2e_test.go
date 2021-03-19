package telegram

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/internal/tgtest"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

type quorumSetup struct {
	TB     testing.TB
	Quorum *tgtest.Quorum
	Logger *zap.Logger
}

type clientSetup struct {
	TB       testing.TB
	Options  Options
	Complete func()
}

func testQuorum(
	trp Transport,
	setup func(q quorumSetup),
	run func(ctx context.Context, c clientSetup) error,
) func(t *testing.T) {
	return func(t *testing.T) {
		log := zaptest.NewLogger(t)
		defer func() { _ = log.Sync() }()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		grp := tdsync.NewCancellableGroup(ctx)

		q := tgtest.NewQuorum(trp.Codec).WithLogger(log.Named("quorum"))
		setup(quorumSetup{
			TB:     t,
			Quorum: q,
			Logger: log,
		})
		grp.Go(q.Up)

		grp.Go(func(ctx context.Context) error {
			select {
			// Await quorum readiness.
			case <-q.Ready():
			case <-ctx.Done():
				return ctx.Err()
			}

			return run(ctx, clientSetup{
				TB: t,
				Options: Options{
					PublicKeys:     q.Keys(),
					Resolver:       dcs.PlainResolver(dcs.PlainOptions{Transport: trp}),
					Logger:         log.Named("client"),
					SessionStorage: &session.StorageMemory{},
					DCList:         q.Config().DCOptions,
				},
				Complete: cancel,
			})
		})

		log.Debug("Waiting")
		if err := grp.Wait(); !errors.Is(err, context.Canceled) {
			require.NoError(t, err)
		}
	}
}

func testAllTransports(t *testing.T, test func(trp Transport) func(t *testing.T)) {
	t.Run("Abridged", test(transport.Abridged()))
	t.Run("Intermediate", test(transport.Intermediate()))
	t.Run("PaddedIntermediate", test(transport.PaddedIntermediate()))
	t.Run("Full", test(transport.Full()))
}

func testTransport(trp Transport) func(t *testing.T) {
	testMessage := "ну че там с деньгами?"

	return testQuorum(trp, func(s quorumSetup) {
		q := s.Quorum

		h := tgtest.TestTransport(s.TB, s.Logger.Named("handler"), testMessage)
		q.Common().Vector(tg.UsersGetUsersRequestTypeID, &tg.User{
			ID:         10,
			AccessHash: 10,
			Username:   "rustcocks",
		})
		q.Dispatch(2, "server").
			Handle(tg.InvokeWithLayerRequestTypeID, h).
			Handle(tg.MessagesSendMessageRequestTypeID, h)
	}, func(ctx context.Context, c clientSetup) error {
		opts := c.Options
		opts.AckBatchSize = 1
		opts.AckInterval = time.Millisecond * 50
		opts.RetryInterval = time.Millisecond * 50
		logger := opts.Logger

		dispatcher := tg.NewUpdateDispatcher()
		opts.UpdateHandler = dispatcher
		client := NewClient(1, "hash", opts)

		waitForMessage := make(chan struct{})
		dispatcher.OnNewMessage(func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
			message := update.Message.(*tg.Message).Message
			logger.Info("Got message", zap.String("text", message))
			assert.Equal(c.TB, testMessage, message)
			if err := client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
				Peer:    &tg.InputPeerUser{},
				Message: "какими деньгами?",
			}); err != nil {
				return err
			}

			logger.Info("Closing waitForMessage")
			close(waitForMessage)
			return nil
		})

		return client.Run(ctx, func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				c.TB.Error("Failed to wait for message")
				return ctx.Err()
			case <-waitForMessage:
				logger.Info("Returning")
				c.Complete()
				return nil
			}
		})
	})
}

func TestClientE2E(t *testing.T) {
	testAllTransports(t, testTransport)
}

func testMigrate(trp Transport) func(t *testing.T) {
	wait := make(chan struct{}, 1)
	return testQuorum(trp, func(s quorumSetup) {
		q := s.Quorum
		q.Common().Vector(tg.UsersGetUsersRequestTypeID, &tg.User{
			ID:         10,
			AccessHash: 10,
			Username:   "rustcocks",
		})
		q.Dispatch(1, "server").HandleFunc(tg.MessagesSendMessageRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				m := &tg.MessagesSendMessageRequest{}
				if err := m.Decode(req.Buf); err != nil {
					return err
				}

				select {
				case wait <- struct{}{}:
				case <-req.RequestCtx.Done():
					return req.RequestCtx.Err()
				}
				return server.SendResult(req, &tg.Updates{})
			},
		)
		q.Dispatch(2, "migrate").HandleFunc(tg.MessagesSendMessageRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				m := &tg.MessagesSendMessageRequest{}
				if err := m.Decode(req.Buf); err != nil {
					return err
				}

				return server.SendResult(req, &mt.RPCError{
					ErrorCode:    303,
					ErrorMessage: "NETWORK_MIGRATE_1",
				})
			},
		)
	}, func(ctx context.Context, c clientSetup) error {
		client := NewClient(1, "hash", c.Options)
		return client.Run(ctx, func(ctx context.Context) error {
			if err := client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
				Peer:    &tg.InputPeerUser{},
				Message: "abc",
			}); err != nil {
				return xerrors.Errorf("send: %w", err)
			}

			select {
			case <-wait:
				c.Complete()
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	})
}

func TestMigrate(t *testing.T) {
	t.Run("Intermediate", testMigrate(transport.Intermediate()))
}
