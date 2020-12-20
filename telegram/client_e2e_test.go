package telegram

import (
	"context"
	"crypto/rsa"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/crypto"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/telegram/internal/tgtest"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

type handler struct {
	server *tgtest.Server
	t      *testing.T

	// For ACK testing proposes.
	// We send ack only after second request
	counter   int
	counterMx sync.Mutex

	message string // immutable
}

func (h *handler) OnNewClient(k crypto.AuthKeyWithID) error {
	h.t.Log("new client connected")

	return nil
}

func (h *handler) hello(k tgtest.Session, message string) error {
	h.t.Log("[server]", "sent message", message)

	return h.server.Send(k, &tg.Updates{
		Updates: []tg.UpdateClass{
			&tg.UpdateNewMessage{
				Message: &tg.Message{
					ID:      1,
					PeerID:  &tg.PeerUser{UserID: 1},
					Message: message,
				},
			},
		},
		Date: int(time.Now().Unix()),
	})
}

func (h *handler) sendConfig(k tgtest.Session, id int64) error {
	return h.server.SendResult(k, id, &tg.Config{})
}

func (h *handler) OnMessage(k tgtest.Session, msgID int64, in *bin.Buffer) error {
	id, err := in.PeekID()
	if err != nil {
		return err
	}

	h.t.Logf("new message, type %x", id)

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

		if err := h.sendConfig(k, msgID); err != nil {
			return err
		}

		return h.hello(k, h.message)
	case tg.MessagesSendMessageRequestTypeID:
		m := &tg.MessagesSendMessageRequest{}
		if err := m.Decode(in); err != nil {
			h.t.Fail()
			return err
		}

		return h.handleMessage(k, msgID, m)
	default:
		h.t.Logf("unexpected type: %x", id)
	}

	return nil
}

func (h *handler) handleMessage(k tgtest.Session, msgID int64, m *tg.MessagesSendMessageRequest) error {
	require.Equal(h.t, "какими деньгами?", m.Message)

	h.counterMx.Lock()
	h.counter++
	if h.counter < 2 {
		h.counterMx.Unlock()
		return nil
	}
	h.counterMx.Unlock()

	updates := &tg.Updates{}
	return h.server.SendResult(k, msgID, updates)
}

func testTransport(trp *transport.Transport) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
		defer cancel()

		srv := tgtest.NewUnstartedServer(ctx, t, trp.Codec())
		h := &handler{
			server:  srv,
			t:       t,
			message: "ну как там с деньгами?",
		}
		srv.SetHandler(h)
		srv.Start()
		defer srv.Close()

		dispatcher := tg.NewUpdateDispatcher()
		log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.DebugLevel))
		client := NewClient(1, "hash", Options{
			PublicKeys:    []*rsa.PublicKey{srv.Key()},
			Addr:          srv.Addr().String(),
			Transport:     trp,
			Logger:        log,
			UpdateHandler: dispatcher.Handle,
		})
		raw := tg.NewClient(client)

		wait := make(chan struct{})
		dispatcher.OnNewMessage(func(uctx tg.UpdateContext, update *tg.UpdateNewMessage) error {
			message := update.Message.(*tg.Message).Message
			t.Log("[client]", "got message", message)
			if message != h.message {
				t.Fatalf("expected %s, got %s", h.message, message)
			}

			_, err := raw.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
				Peer:    &tg.InputPeerUser{},
				Message: "какими деньгами?",
			})
			if err != nil {
				return err
			}

			wait <- struct{}{}
			return client.Close(ctx)
		})

		err := client.Connect(ctx)
		if err != nil {
			t.Fatal(err)
		}

		<-wait
	}
}

func TestClient(t *testing.T) {
	t.Run("intermediate", testTransport(transport.Intermediate(nil)))
	t.Run("full", testTransport(transport.Full(nil)))
}
