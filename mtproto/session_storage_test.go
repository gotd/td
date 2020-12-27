package mtproto

import (
	"context"
	"io"
	"testing"

	"github.com/gotd/td/internal/testutil"

	"github.com/stretchr/testify/require"
)

type testStorage struct {
	load  func() ([]byte, error)
	store func([]byte) error
}

func (s testStorage) LoadSession(ctx context.Context) ([]byte, error)     { return s.load() }
func (s testStorage) StoreSession(ctx context.Context, data []byte) error { return s.store(data) }

func TestClientSessionStorage(t *testing.T) {
	client := newTestClient(nil)
	ctx := context.Background()
	t.Run("Ok", func(t *testing.T) {
		var sessionRaw []byte
		client.sessionStorage = testStorage{
			load: func() ([]byte, error) {
				if len(sessionRaw) == 0 {
					return nil, ErrSessionNotFound
				}
				return sessionRaw, nil
			},
			store: func(data []byte) error {
				sessionRaw = data
				return nil
			},
		}
		require.NoError(t, client.loadSession(ctx), "should ignore ErrSessionNotFound")
		require.NoError(t, client.saveSession(ctx), "should save session")
		require.NoError(t, client.loadSession(ctx), "should load session")
	})
	t.Run("Error", func(t *testing.T) {
		expectedErr := io.ErrClosedPipe
		client.sessionStorage = testStorage{
			load: func() ([]byte, error) {
				return nil, expectedErr
			},
			store: func(bytes []byte) error {
				return expectedErr
			},
		}
		testutil.RequireErr(t, expectedErr, client.loadSession(ctx))
		testutil.RequireErr(t, expectedErr, client.saveSession(ctx))
	})
}
