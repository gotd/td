package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

type mockHandler struct {
	LastUpdate      *tg.Updates
	LastShortUpdate *tg.UpdateShort
}

func (m *mockHandler) Handle(ctx context.Context, u *tg.Updates) error {
	m.LastUpdate = u
	return nil
}

func (m *mockHandler) HandleShort(ctx context.Context, u *tg.UpdateShort) error {
	m.LastShortUpdate = u
	return nil
}

func TestClient_processUpdates(t *testing.T) {
	t.Run("Nil handler", func(t *testing.T) {
		c := new(Client)
		require.NoError(t, c.processUpdates(nil))
	})

	msg := &tg.Message{
		ID: 1,
	}
	upd := &tg.UpdateNewMessage{
		Message: msg,
	}

	t.Run("Handle", func(t *testing.T) {
		mock := &mockHandler{}
		c := new(Client)
		c.updateHandler = mock

		err := c.processUpdates(&tg.Updates{Updates: []tg.UpdateClass{upd}})
		require.NoError(t, err)
		require.Equal(t, upd, mock.LastUpdate.Updates[0])
	})

	t.Run("HandleShort", func(t *testing.T) {
		mock := &mockHandler{}
		c := new(Client)
		c.updateHandler = mock

		err := c.processUpdates(&tg.UpdateShort{Update: upd})
		require.NoError(t, err)
		require.Equal(t, upd, mock.LastShortUpdate.Update)

		err = c.processUpdates(&tg.UpdateShortMessage{ID: 10})
		require.NoError(t, err)
		require.Equal(t, 10, mock.LastShortUpdate.Update.(*tg.UpdateNewMessage).Message.(*tg.Message).ID)

		err = c.processUpdates(&tg.UpdateShortSentMessage{ID: 10})
		require.NoError(t, err)
		require.Equal(t, 10, mock.LastShortUpdate.Update.(*tg.UpdateNewMessage).Message.(*tg.Message).ID)

		err = c.processUpdates(&tg.UpdateShortChatMessage{ID: 10})
		require.NoError(t, err)
		require.Equal(t, 10, mock.LastShortUpdate.Update.(*tg.UpdateNewMessage).Message.(*tg.Message).ID)
	})
}
