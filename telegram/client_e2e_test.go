package telegram_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/tdsync"
	"github.com/nnqq/td/session"
	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/dcs"
	"github.com/nnqq/td/telegram/downloader"
	"github.com/nnqq/td/telegram/uploader"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
	"github.com/nnqq/td/tgtest"
	"github.com/nnqq/td/tgtest/cluster"
	"github.com/nnqq/td/tgtest/services/file"
	"github.com/nnqq/td/transport"
)

type clusterSetup struct {
	TB      testing.TB
	Cluster *cluster.Cluster
	Logger  *zap.Logger
}

type clientSetup struct {
	TB       testing.TB
	Options  telegram.Options
	Complete func()
}

var user = &tg.User{
	ID:         10,
	AccessHash: 10,
	Username:   "username",
}

func testCluster(
	p dcs.Protocol,
	ws bool,
	setup func(q clusterSetup),
	run func(ctx context.Context, c clientSetup) error,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		log := zaptest.NewLogger(t)
		defer func() { _ = log.Sync() }()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		g := tdsync.NewCancellableGroup(ctx)

		c := cluster.NewCluster(cluster.Options{
			Web:      ws,
			Logger:   log.Named("cluster"),
			Protocol: p,
		})
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
					UpdateHandler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
						// No-op update handler.
						return nil
					}),
					PublicKeys:     c.Keys(),
					Resolver:       c.Resolver(),
					Logger:         log.Named("client"),
					SessionStorage: &session.StorageMemory{},
					DCList:         c.List(),
				},
				Complete: cancel,
			})
		})

		log.Debug("Waiting")
		if err := g.Wait(); err != nil && !xerrors.Is(err, context.Canceled) {
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

	return testCluster(p, false, func(s clusterSetup) {
		h := tgtest.TestTransport(s.TB, s.Logger.Named("handler"), testMessage)
		d := s.Cluster.Dispatch(2, "server")
		d.Handle(tg.MessagesSendMessageRequestTypeID, h)
		d.Handle(tg.UsersGetUsersRequestTypeID, h)
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
	test := func(ws bool) func(t *testing.T) {
		wait := make(chan struct{}, 1)
		return testCluster(p, ws, func(s clusterSetup) {
			c := s.Cluster
			c.Common().Vector(tg.UsersGetUsersRequestTypeID, user)
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
					return server.SendGZIP(req, &tg.Updates{})
				},
			)
			c.Dispatch(2, "migrate").HandleFunc(tg.MessagesSendMessageRequestTypeID,
				func(server *tgtest.Server, req *tgtest.Request) error {
					m := &tg.MessagesSendMessageRequest{}
					if err := m.Decode(req.Buf); err != nil {
						return err
					}

					return server.SendErr(req, tgerr.New(303, "NETWORK_MIGRATE_1"))
				},
			)
		}, func(ctx context.Context, c clientSetup) error {
			opts := c.Options
			opts.MigrationTimeout = time.Minute
			client := telegram.NewClient(1, "hash", opts)
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

	return func(t *testing.T) {
		t.Run("TCP", test(false))
		t.Run("Websocket", test(true))
	}
}

func TestMigrate(t *testing.T) {
	t.Run("Intermediate", testMigrate(transport.Intermediate))
}

func testFiles(p dcs.Protocol) func(t *testing.T) {
	test := func(ws bool) func(t *testing.T) {
		return testCluster(p, ws, func(s clusterSetup) {
			c := s.Cluster
			c.Common().Vector(tg.UsersGetUsersRequestTypeID, user)
			f := file.NewService(file.Config{
				HashPartSize: 1024,
			})
			f.Register(c.Dispatch(2, "DC"))
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

	return func(t *testing.T) {
		t.Run("TCP", test(false))
		t.Run("Websocket", test(true))
	}
}

func TestFiles(t *testing.T) {
	t.Run("Intermediate", testFiles(transport.Intermediate))
}
