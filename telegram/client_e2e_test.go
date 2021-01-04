package telegram

import (
	"context"
	"crypto/rsa"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram/internal/tgtest"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

func testTransport(trp Transport) func(t *testing.T) {
	return func(t *testing.T) {
		log := zaptest.NewLogger(t)
		defer func() { _ = log.Sync() }()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		g, gctx := errgroup.WithContext(ctx)

		testMessage := "ну че там с деньгами?"
		suite := tgtest.NewSuite(gctx, t, log)
		srv := tgtest.TestTransport(suite, testMessage, trp.Codec)
		g.Go(func() error {
			defer srv.Close()
			return srv.Serve()
		})

		g.Go(func() error {
			dispatcher := tg.NewUpdateDispatcher()
			clientLogger := log.Named("client")
			client := NewClient(1, "hash", Options{
				PublicKeys:     []*rsa.PublicKey{srv.Key()},
				Addr:           srv.Addr().String(),
				Transport:      trp,
				Logger:         clientLogger,
				UpdateHandler:  dispatcher.Handle,
				AckBatchSize:   1,
				AckInterval:    time.Millisecond * 50,
				RetryInterval:  time.Millisecond * 50,
				SessionStorage: &session.StorageMemory{},
			})

			waitForMessage := make(chan struct{})
			dispatcher.OnNewMessage(func(uctx tg.UpdateContext, update *tg.UpdateNewMessage) error {
				message := update.Message.(*tg.Message).Message
				clientLogger.With(zap.String("message", message)).
					Info("got message")
				require.Equal(t, testMessage, message)

				err := client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
					Peer:    &tg.InputPeerUser{},
					Message: "какими деньгами?",
				})
				if err != nil {
					return err
				}

				close(waitForMessage)
				return nil
			})

			defer srv.Close() // stop server in any case
			if err := client.Connect(ctx); err != nil {
				return err
			}
			<-waitForMessage

			return client.Close()
		})

		if err := g.Wait(); !errors.Is(err, context.Canceled) {
			require.NoError(t, err)
		}
	}
}

func TestClient(t *testing.T) {
	t.Run("Abridged", testTransport(transport.Abridged(nil)))
	t.Run("Intermediate", testTransport(transport.Intermediate(nil)))
	t.Run("PaddedIntermediate", testTransport(transport.PaddedIntermediate(nil)))
	t.Run("Full", testTransport(transport.Full(nil)))
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
		t.Helper()
		log := zaptest.NewLogger(t)
		defer func() { _ = log.Sync() }()

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
		defer cancel()

		srv := tgtest.NewUnstartedServer(tgtest.NewSuite(ctx, t, log), trp.Codec)

		alreadyConnected := newSyncHashSet()
		wait := make(chan struct{})
		srv.SetHandlerFunc(func(s tgtest.Session, msgID int64, in *bin.Buffer) error {
			id, err := in.PeekID()
			if err != nil {
				return err
			}

			switch id {
			case proto.InvokeWithLayerID:
				layerInvoke := proto.InvokeWithLayer{
					Query: &proto.InitConnection{
						Query: proto.GetConfig{},
					},
				}

				if err := layerInvoke.Decode(in); err != nil {
					return err
				}

				return srv.SendConfig(s, msgID)
			case mt.PingDelayDisconnectRequestTypeID:
				pingReq := mt.PingDelayDisconnectRequest{}
				if err := pingReq.Decode(in); err != nil {
					return err
				}

				return srv.SendPong(s, msgID, pingReq.PingID)
			case tg.MessagesSendMessageRequestTypeID:
				m := &tg.MessagesSendMessageRequest{}
				if err := m.Decode(in); err != nil {
					return err
				}
				require.Equal(t, testMessage, m.Message)

				if alreadyConnected.Has(s.AuthKeyID) {
					srv.ForceDisconnect(s)
					alreadyConnected.Add(s.AuthKeyID)
					return nil
				}

				wait <- struct{}{}
				return srv.SendResult(s, msgID, &tg.Updates{})
			}

			return nil
		})
		srv.Start()
		defer srv.Close()

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

		err := client.Connect(ctx)
		if err != nil {
			t.Fatal(err)
		}

		_ = client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
			Peer:    &tg.InputPeerUser{},
			Message: testMessage,
		})

		<-wait
	}
}

func TestReconnect(t *testing.T) {
	if os.Getenv("GOTD_TEST_RECONNECT") != "1" {
		t.Skip("TODO: Fix flaky test")
	}

	t.Run("intermediate", testReconnect(transport.Intermediate(nil)))
	t.Run("full", testReconnect(transport.Full(nil)))
}
