package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

type mockHandler struct {
	LastUpdate tg.UpdatesClass
}

func (m *mockHandler) Handle(ctx context.Context, u tg.UpdatesClass) error {
	m.LastUpdate = u
	return nil
}

func TestClient_processUpdates(t *testing.T) {
	msg := &tg.Message{
		ID: 1,
	}
	upd := &tg.Updates{
		Updates: []tg.UpdateClass{&tg.UpdateNewMessage{
			Message: msg,
		}},
	}

	t.Run("Handle", func(t *testing.T) {
		mock := &mockHandler{}
		c := new(Client)
		c.updateHandler = mock

		err := c.processUpdates(upd)
		require.NoError(t, err)
		require.Equal(t, upd, mock.LastUpdate)
	})
}
