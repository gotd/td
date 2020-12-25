package tgtest

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

type testTransportHandler struct {
	server *Server
	t      testing.TB
	logger *zap.Logger

	// For ACK testing proposes.
	// We send ack only after second request
	counter   int
	counterMx sync.Mutex

	message string // immutable
	types   *tmap.Map
}

func TestTransport(s Suite, message string, codec func() transport.Codec) *Server {
	srv := NewUnstartedServer(s, codec)
	h := testTransport(s, srv, message)
	srv.SetHandler(h)

	return srv
}

func testTransport(s Suite, server *Server, message string) *testTransportHandler {
	return &testTransportHandler{
		server:  server,
		t:       s.TB,
		logger:  s.Log.Named("handler"),
		message: message,
		types: tmap.New(
			mt.TypesMap(),
			tg.TypesMap(),
			proto.TypesMap(),
		),
	}
}

func (h *testTransportHandler) hello(k Session, message string) error {
	h.logger.With(zap.String("message", message)).
		Info("sent message")

	return h.server.SendUpdates(k, &tg.UpdateNewMessage{
		Message: &tg.Message{
			ID:      1,
			PeerID:  &tg.PeerUser{UserID: 1},
			Message: message,
		},
	})
}

func (h *testTransportHandler) OnMessage(k Session, msgID int64, in *bin.Buffer) error {
	id, err := in.PeekID()
	if err != nil {
		return err
	}

	h.logger.With(
		zap.String("id", fmt.Sprintf("%x", id)),
	).Info("new message")

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
		h.logger.Info("new client connected, invoke received")

		if err := h.server.SendConfig(k, msgID); err != nil {
			return err
		}

		return h.hello(k, h.message)
	case mt.PingDelayDisconnectRequestTypeID:
		pingReq := mt.PingDelayDisconnectRequest{}
		if err := pingReq.Decode(in); err != nil {
			return err
		}

		return h.server.SendPong(k, msgID, pingReq.PingID)
	case tg.MessagesSendMessageRequestTypeID:
		m := &tg.MessagesSendMessageRequest{}
		if err := m.Decode(in); err != nil {
			h.t.Fail()
			return err
		}

		return h.handleMessage(k, msgID, m)
	}

	return nil
}

func (h *testTransportHandler) handleMessage(k Session, msgID int64, m *tg.MessagesSendMessageRequest) error {
	require.Equal(h.t, "какими деньгами?", m.Message)

	h.counterMx.Lock()
	h.counter++
	if h.counter < 2 {
		h.counterMx.Unlock()
		return nil
	}
	h.counterMx.Unlock()

	return h.server.SendResult(k, msgID, &tg.Updates{})
}
