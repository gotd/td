package mtproto

import (
	"context"
	"io"
	"testing"

	"github.com/gotd/td/session"

	"github.com/gotd/td/internal/testutil"

	"github.com/stretchr/testify/require"
)

type testStorage struct {
	load  func() (*session.Data, error)
	store func(*session.Data) error
}

func (s testStorage) Load(ctx context.Context) (*session.Data, error) { return s.load() }
func (s testStorage) Save(ctx context.Context, data *session.Data) error {
	return s.store(data)
}

func TestClientSessionStorage(t *testing.T) {
	client := newTestClient(nil)
	ctx := context.Background()
	t.Run("Ok", func(t *testing.T) {
		var s *session.Data
		client.session = testStorage{
			load: func() (*session.Data, error) {
				if s == nil {
					return nil, session.ErrNotFound
				}
				return s, nil
			},
			store: func(data *session.Data) error {
				s = data
				return nil
			},
		}
		require.NoError(t, client.loadSession(ctx), "should ignore ErrSessionNotFound")
		require.NoError(t, client.saveSession(ctx), "should save session")
		require.NoError(t, client.loadSession(ctx), "should load session")
	})
	t.Run("Error", func(t *testing.T) {
		expectedErr := io.ErrClosedPipe
		client.session = testStorage{
			load: func() (*session.Data, error) {
				return nil, expectedErr
			},
			store: func(*session.Data) error {
				return expectedErr
			},
		}
		testutil.RequireErr(t, expectedErr, client.loadSession(ctx))
		testutil.RequireErr(t, expectedErr, client.saveSession(ctx))
	})
}
