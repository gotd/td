package telegram

import (
	"context"
	"crypto/rsa"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram/internal/tgtest"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

func testAllTransports(t *testing.T, test func(trp Transport) func(t *testing.T)) {
	t.Run("Abridged", test(transport.Abridged(nil)))
	t.Run("Intermediate", test(transport.Intermediate(nil)))
	t.Run("PaddedIntermediate", test(transport.PaddedIntermediate(nil)))
	t.Run("Full", test(transport.Full(nil)))
}

func testTransport(trp Transport) func(t *testing.T) {
	return func(t *testing.T) {
		log := zaptest.NewLogger(t)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		g, gCtx := errgroup.WithContext(ctx)

		testMessage := "ну че там с деньгами?"
		suite := tgtest.NewSuite(gCtx, t, log)
		srv := tgtest.TestTransport(suite, testMessage, trp.Codec)
		g.Go(func() error {
			defer srv.Close()
			return srv.Serve()
		})
		g.Go(func() error {
			dispatcher := tg.NewUpdateDispatcher()
			logger := log.Named("client")
			client := NewClient(1, "hash", Options{
				PublicKeys:     []*rsa.PublicKey{srv.Key()},
				Addr:           srv.Addr().String(),
				Transport:      trp,
				Logger:         logger,
				UpdateHandler:  dispatcher,
				AckBatchSize:   1,
				AckInterval:    time.Millisecond * 50,
				RetryInterval:  time.Millisecond * 50,
				SessionStorage: &session.StorageMemory{},
			})

			waitForMessage := make(chan struct{})
			dispatcher.OnNewMessage(func(ctx tg.UpdateContext, update *tg.UpdateNewMessage) error {
				message := update.Message.(*tg.Message).Message
				logger.Info("Got message", zap.String("text", message))
				assert.Equal(t, testMessage, message)
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

			return client.Run(gCtx, func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					t.Error("Failed to wait for message")
					return ctx.Err()
				case <-waitForMessage:
					logger.Info("Returning")
					cancel()
					return nil
				}
			})
		})

		log.Debug("Waiting")
		if err := g.Wait(); !errors.Is(err, context.Canceled) {
			require.NoError(t, err)
		}
	}
}

func TestClientE2E(t *testing.T) {
	testAllTransports(t, testTransport)
}

func defaultMigrationHandler(dcOps *tg.Config) tgtest.HandlerFunc {
	return func(srv *tgtest.Server, req *tgtest.Request) error {
		id, err := req.Buf.PeekID()
		if err != nil {
			return err
		}

		switch id {
		case tg.InvokeWithLayerRequestTypeID:
			layerInvoke := tg.InvokeWithLayerRequest{
				Query: &tg.InitConnectionRequest{
					Query: &tg.HelpGetConfigRequest{},
				},
			}

			if err := layerInvoke.Decode(req.Buf); err != nil {
				return err
			}

			return srv.SendResult(req, dcOps)
		case tg.UsersGetUsersRequestTypeID:
			return srv.SendVector(req, &tg.User{
				ID:         10,
				AccessHash: 10,
				Username:   "jit_rs",
			})
		case tg.HelpGetConfigRequestTypeID:
			return srv.SendResult(req, dcOps)
		default:
			return xerrors.Errorf("unexpected TypeID %x call", id)
		}
	}
}

func testMigrate(trp Transport) func(t *testing.T) {
	return func(t *testing.T) {
		log := zaptest.NewLogger(t)
		defer func() { _ = log.Sync() }()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		g, gCtx := errgroup.WithContext(ctx)
		suite := tgtest.NewSuite(gCtx, t, log)

		srv := tgtest.NewUnstartedServer("server", suite, trp.Codec)
		migrate := tgtest.NewUnstartedServer("migrate", suite, trp.Codec)

		srvAddr, ok := srv.Addr().(*net.TCPAddr)
		require.Truef(t, ok, "unexpected type %T", srv.Addr())
		migrateAddr, ok := migrate.Addr().(*net.TCPAddr)
		require.Truef(t, ok, "unexpected type %T", migrate.Addr())
		dcOps := &tg.Config{
			DCOptions: []tg.DCOption{
				{
					ID:        1,
					IPAddress: srvAddr.IP.String(),
					Port:      srvAddr.Port,
				},
				{
					ID:        2,
					IPAddress: migrateAddr.IP.String(),
					Port:      migrateAddr.Port,
				},
			},
		}

		wait := make(chan struct{})
		srv.Dispatcher().HandleFunc(tg.MessagesSendMessageRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				m := &tg.MessagesSendMessageRequest{}
				if err := m.Decode(req.Buf); err != nil {
					return err
				}

				wait <- struct{}{}
				return srv.SendResult(req, &tg.Updates{})
			},
		).Fallback(defaultMigrationHandler(dcOps))
		g.Go(func() error {
			defer srv.Close()
			return srv.Serve()
		})

		migrate.Dispatcher().HandleFunc(tg.MessagesSendMessageRequestTypeID,
			func(server *tgtest.Server, req *tgtest.Request) error {
				m := &tg.MessagesSendMessageRequest{}
				if err := m.Decode(req.Buf); err != nil {
					return err
				}

				return migrate.SendResult(req, &mt.RPCError{
					ErrorCode:    303,
					ErrorMessage: "NETWORK_MIGRATE_1",
				})
			}).Fallback(defaultMigrationHandler(dcOps))
		g.Go(func() error {
			defer migrate.Close()
			return migrate.Serve()
		})

		g.Go(func() error {
			client := NewClient(1, "hash", Options{
				PublicKeys:     []*rsa.PublicKey{migrate.Key(), srv.Key()},
				Addr:           migrate.Addr().String(),
				Transport:      trp,
				Logger:         log.Named("client"),
				SessionStorage: &session.StorageMemory{},
			})

			return client.Run(gCtx, func(ctx context.Context) error {
				if err := client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
					Peer:    &tg.InputPeerUser{},
					Message: "abc",
				}); err != nil {
					return xerrors.Errorf("send: %w", err)
				}

				return nil
			})
		})
		g.Go(func() error {
			select {
			case <-wait:
				cancel()
				return nil
			case <-gCtx.Done():
				t.Error("failed to wait")
				return gCtx.Err()
			}
		})

		log.Debug("Waiting")
		if err := g.Wait(); !errors.Is(err, context.Canceled) {
			require.NoError(t, err)
		}
	}
}

func TestMigrate(t *testing.T) {
	t.Run("Intermediate", testMigrate(transport.Intermediate(nil)))
}
