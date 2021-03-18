package tgtest

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

type testTransportHandler struct {
	t      testing.TB
	logger *zap.Logger
	// For ACK testing proposes.
	// We send ack only after second request
	counter   int
	counterMx sync.Mutex

	message string // immutable
}

func TestTransport(t testing.TB, logger *zap.Logger, message string) Handler {
	return &testTransportHandler{
		t:       t,
		logger:  logger,
		message: message,
	}
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

		if err := server.SendConfig(req); err != nil {
			return err
		}

		h.logger.Info("Sending message", zap.String("message", h.message))
		return server.SendUpdates(req.RequestCtx, req.Session, &tg.UpdateNewMessage{
			Message: &tg.Message{
				ID:      1,
				PeerID:  &tg.PeerUser{UserID: 1},
				Message: h.message,
			},
		})
	case tg.MessagesSendMessageRequestTypeID:
		m := &tg.MessagesSendMessageRequest{}
		if err := m.Decode(req.Buf); err != nil {
			h.t.Fail()
			return err
		}

		require.Equal(h.t, "какими деньгами?", m.Message)

		h.counterMx.Lock()
		h.counter++
		if h.counter < 2 {
			h.counterMx.Unlock()
			return nil
		}
		h.counterMx.Unlock()

		return server.SendResult(req, &tg.Updates{})
	}

	return nil
}
