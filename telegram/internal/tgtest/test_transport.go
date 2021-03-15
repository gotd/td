package tgtest

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

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
	srv := NewUnstartedServer("server", s, codec)
	h := testTransport(s, srv, message)
	srv.Dispatcher().Fallback(h)

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
	h.logger.Info("Sent message", zap.String("message", message))

	return h.server.SendUpdates(k, &tg.UpdateNewMessage{
		Message: &tg.Message{
			ID:      1,
			PeerID:  &tg.PeerUser{UserID: 1},
			Message: message,
		},
	})
}

func (h *testTransportHandler) OnMessage(server *Server, req *Request) error {
	id, err := req.Buf.PeekID()
	if err != nil {
		return err
	}

	h.logger.Info("New message", zap.String("id", fmt.Sprintf("%x", id)))

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
		h.logger.Info("New client connected, invoke received")

		if err := h.server.SendConfig(req); err != nil {
			return err
		}

		return h.hello(req.Session, h.message)
	case tg.MessagesSendMessageRequestTypeID:
		m := &tg.MessagesSendMessageRequest{}
		if err := m.Decode(req.Buf); err != nil {
			h.t.Fail()
			return err
		}

		return h.handleMessage(req, m)
	}

	return nil
}

func (h *testTransportHandler) handleMessage(req *Request, m *tg.MessagesSendMessageRequest) error {
	require.Equal(h.t, "какими деньгами?", m.Message)

	h.counterMx.Lock()
	h.counter++
	if h.counter < 2 {
		h.counterMx.Unlock()
		return nil
	}
	h.counterMx.Unlock()

	return h.server.SendResult(req, &tg.Updates{})
}
