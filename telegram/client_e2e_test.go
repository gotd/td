package telegram

import (
	"context"
	"crypto/rsa"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"

	"github.com/stretchr/testify/require"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gotd/td/telegram/internal/tgtest"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

func testTransport(trp Transport) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.DebugLevel))

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
		defer cancel()

		testMessage := "ну че там с деньгами?"
		suite := tgtest.NewSuite(ctx, t, log)
		srv := tgtest.TestTransport(suite, testMessage, trp.Codec)
		srv.Start()
		defer srv.Close()

		dispatcher := tg.NewUpdateDispatcher()
		clientLogger := log.Named("client")
		client := NewClient(1, "hash", Options{
			PublicKeys:    []*rsa.PublicKey{srv.Key()},
			Addr:          srv.Addr().String(),
			Transport:     trp,
			Logger:        clientLogger,
			UpdateHandler: dispatcher.Handle,
			AckBatchSize:  1,
			AckInterval:   time.Millisecond * 50,
			RetryInterval: time.Millisecond * 50,
		})

		wait := make(chan struct{})
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

			wait <- struct{}{}
			return client.Close()
		})

		err := client.Connect(ctx)
		if err != nil {
			t.Fatal(err)
		}

		<-wait
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
		log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.DebugLevel))

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
		defer cancel()

		srv := tgtest.NewUnstartedServer(ctx, trp.Codec)

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
