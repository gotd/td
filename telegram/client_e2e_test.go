package telegram

import (
	"context"
	"crypto/rsa"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
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
		g, ctx := errgroup.WithContext(ctx)

		testMessage := "ну че там с деньгами?"
		suite := tgtest.NewSuite(ctx, t, log)
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

			return client.Run(ctx, func(ctx context.Context) error {
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

type syncHashSet struct {
	set map[[8]byte]struct{}
	m   sync.Mutex
}

func newSyncHashSet() *syncHashSet {
	return &syncHashSet{set: map[[8]byte]struct{}{}}
}

func (s *syncHashSet) Add(k [8]byte) {
	s.m.Lock()
	s.set[k] = struct{}{}
	s.m.Unlock()
}

func (s *syncHashSet) Has(k [8]byte) (ok bool) {
	s.m.Lock()
	_, ok = s.set[k]
	s.m.Unlock()
	return
}

func testReconnect(trp Transport) func(t *testing.T) {
	testMessage := "какими деньгами?"
	return func(t *testing.T) {
		log := zaptest.NewLogger(t)
		defer func() { _ = log.Sync() }()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		g, gCtx := errgroup.WithContext(ctx)

		srv := tgtest.NewUnstartedServer("server", tgtest.NewSuite(gCtx, t, log), trp.Codec)
		alreadyConnected := newSyncHashSet()
		wait := make(chan struct{})
		srv.SetHandlerFunc(func(s tgtest.Session, msgID int64, in *bin.Buffer) error {
			id, err := in.PeekID()
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

				if err := layerInvoke.Decode(in); err != nil {
					return err
				}

				return srv.SendConfig(s, msgID)
			case tg.HelpGetConfigRequestTypeID:
				return srv.SendConfig(s, msgID)
			case tg.MessagesSendMessageRequestTypeID:
				m := &tg.MessagesSendMessageRequest{}
				if err := m.Decode(in); err != nil {
					return err
				}
				require.Equal(t, testMessage, m.Message)

				if alreadyConnected.Has(s.ID) {
					srv.ForceDisconnect(s)
					alreadyConnected.Add(s.ID)
					return nil
				}

				wait <- struct{}{}
				return srv.SendResult(s, msgID, &tg.Updates{})
			}

			return nil
		})

		g.Go(func() error {
			defer srv.Close()
			return srv.Serve()
		})
		g.Go(func() error {
			client := NewClient(1, "hash", Options{
				PublicKeys:    []*rsa.PublicKey{srv.Key()},
				Addr:          srv.Addr().String(),
				Transport:     trp,
				Logger:        log.Named("client"),
				AckBatchSize:  1,
				AckInterval:   time.Millisecond * 100,
				RetryInterval: time.Millisecond * 100,
				MaxRetries:    5,
			})

			return client.Run(gCtx, func(ctx context.Context) error {
				if err := client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
					Peer:    &tg.InputPeerUser{},
					Message: testMessage,
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

func TestReconnect(t *testing.T) {
	// TODO(ccln): Fix this
	testAllTransports(t, testReconnect)
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

		srv.SetHandlerFunc(func(s tgtest.Session, msgID int64, in *bin.Buffer) error {
			id, err := in.PeekID()
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

				if err := layerInvoke.Decode(in); err != nil {
					return err
				}

				return srv.SendResult(s, msgID, dcOps)
			case tg.HelpGetConfigRequestTypeID:
				return srv.SendResult(s, msgID, dcOps)
			case tg.MessagesSendMessageRequestTypeID:
				m := &tg.MessagesSendMessageRequest{}
				if err := m.Decode(in); err != nil {
					return err
				}

				return srv.SendResult(s, msgID, &tg.Updates{})
			}

			return nil
		})
		g.Go(func() error {
			defer srv.Close()
			return srv.Serve()
		})

		migrate.SetHandlerFunc(func(s tgtest.Session, msgID int64, in *bin.Buffer) error {
			id, err := in.PeekID()
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

				if err := layerInvoke.Decode(in); err != nil {
					return err
				}

				return migrate.SendResult(s, msgID, dcOps)
			case tg.HelpGetConfigRequestTypeID:
				return migrate.SendResult(s, msgID, dcOps)
			case tg.MessagesSendMessageRequestTypeID:
				m := &tg.MessagesSendMessageRequest{}
				if err := m.Decode(in); err != nil {
					return err
				}

				return migrate.SendResult(s, msgID, &mt.RPCError{
					ErrorCode:    303,
					ErrorMessage: "NETWORK_MIGRATE_1",
				})
			default:
				return nil
			}
		})
		g.Go(func() error {
			defer migrate.Close()
			return migrate.Serve()
		})

		gotResponse := make(chan struct{})
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
				close(gotResponse)
				return nil
			})
		})
		g.Go(func() error {
			select {
			case <-gotResponse:
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
