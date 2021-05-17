package telegram_test

import (
	"bytes"
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
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/internal/tgtest"
	"github.com/gotd/td/telegram/internal/tgtest/services/file"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/tgtest/services/file"
	"github.com/gotd/td/transport"
)

type clusterSetup struct {
	TB      testing.TB
	Cluster *tgtest.Cluster
	Logger  *zap.Logger
}

type clientSetup struct {
	TB       testing.TB
	Options  telegram.Options
	Complete func()
}

func testCluster(
	p dcs.Protocol,
	setup func(q clusterSetup),
	run func(ctx context.Context, c clientSetup) error,
) func(t *testing.T) {
	return func(t *testing.T) {
		log := zaptest.NewLogger(t)
		defer func() { _ = log.Sync() }()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		g := tdsync.NewCancellableGroup(ctx)

		c := tgtest.NewCluster(p.Codec).WithLogger(log.Named("cluster"))
		setup(clusterSetup{
			TB:      t,
			Cluster: c,
			Logger:  log,
		})
		g.Go(c.Up)

		g.Go(func(ctx context.Context) error {
			select {
			// Wait for cluster readiness.
			case <-c.Ready():
			case <-ctx.Done():
				return ctx.Err()
			}

			return run(ctx, clientSetup{
				TB: t,
				Options: telegram.Options{
					PublicKeys:     c.Keys(),
					Resolver:       dcs.PlainResolver(dcs.PlainOptions{Protocol: p}),
					Logger:         log.Named("client"),
					SessionStorage: &session.StorageMemory{},
					DCList:         dcs.DCList{
						Options: q.Config().DCOptions,
						Domains: map[int]string{},
					},
				},
				Complete: cancel,
			})
		})

		log.Debug("Waiting")
		if err := g.Wait(); !errors.Is(err, context.Canceled) {
			require.NoError(t, err)
		}
	}
}

func testAllTransports(t *testing.T, test func(p dcs.Protocol) func(t *testing.T)) {
	t.Run("Abridged", test(transport.Abridged))
	t.Run("Intermediate", test(transport.Intermediate))
	t.Run("PaddedIntermediate", test(transport.PaddedIntermediate))
	t.Run("Full", test(transport.Full))
}

func testTransport(p dcs.Protocol) func(t *testing.T) {
	testMessage := "ну че там с деньгами?"

	return testCluster(p, func(s clusterSetup) {
		c := s.Cluster

		h := tgtest.TestTransport(s.TB, s.Logger.Named("handler"), testMessage)
		c.Common().Vector(tg.UsersGetUsersRequestTypeID, &tg.User{
			ID:         10,
			AccessHash: 10,
			Username:   "rustcocks",
		})
		c.Dispatch(2, "server").
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
		client := telegram.NewClient(1, "hash", opts)

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

func testMigrate(p dcs.Protocol) func(t *testing.T) {
	wait := make(chan struct{}, 1)
	return testCluster(p, func(s clusterSetup) {
		c := s.Cluster
		c.Common().Vector(tg.UsersGetUsersRequestTypeID, &tg.User{
			ID:         10,
			AccessHash: 10,
			Username:   "rustcocks",
		})
		c.Dispatch(1, "server").HandleFunc(tg.MessagesSendMessageRequestTypeID,
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
		c.Dispatch(2, "migrate").HandleFunc(tg.MessagesSendMessageRequestTypeID,
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
		client := telegram.NewClient(1, "hash", c.Options)
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
	t.Run("Intermediate", testMigrate(transport.Intermediate))
}

func testFiles(p dcs.Protocol) func(t *testing.T) {
	return testCluster(p, func(s clusterSetup) {
		c := s.Cluster
		c.Common().Vector(tg.UsersGetUsersRequestTypeID, &tg.User{
			ID:         10,
			AccessHash: 10,
			Username:   "rustcocks",
		})
		f := file.NewService(file.NewInMemory()).WitHashPartSize(1024)
		f.Register(c.DC(2, "DC").Dispatcher())
	}, func(ctx context.Context, c clientSetup) error {
		client := telegram.NewClient(1, "hash", c.Options)
		defer c.Complete()
		return client.Run(ctx, func(ctx context.Context) error {
			raw := tg.NewClient(client)
			upd := uploader.NewUploader(raw)
			dwn := downloader.NewDownloader()

			payloads := [][]byte{
				[]byte("data"),
				bytes.Repeat([]byte{10}, 1337),
				bytes.Repeat([]byte{42}, 16384),
			}

			for _, payload := range payloads {
				f, err := upd.FromBytes(ctx, "10.jpg", payload)
				if err != nil {
					return err
				}

				var b bytes.Buffer
				_, err = dwn.Download(raw, &tg.InputFileLocation{
					VolumeID: f.GetID(),
					LocalID:  10,
				}).WithVerify(true).Stream(ctx, &b)
				if err != nil {
					return err
				}

				if !bytes.Equal(payload, b.Bytes()) {
					c.TB.Error("must be equal")
				}
			}
			return nil
		})
	})
}

func TestFiles(t *testing.T) {
	t.Run("Intermediate", testFiles(transport.Intermediate))
}
