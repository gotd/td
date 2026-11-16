package peers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type accessHashContextStorage struct {
	InmemoryStorage
	checkContext func(context.Context)
}

func (s *accessHashContextStorage) Save(ctx context.Context, key Key, value Value) error {
	s.checkContext(ctx)
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.InmemoryStorage.Save(ctx, key, value)
}

func (s *accessHashContextStorage) Find(ctx context.Context, key Key) (Value, bool, error) {
	s.checkContext(ctx)
	if err := ctx.Err(); err != nil {
		return Value{}, false, err
	}
	return s.InmemoryStorage.Find(ctx, key)
}

func TestManager_ChannelAccessHashContext(t *testing.T) {
	for _, method := range []string{"Set", "Get"} {
		t.Run(method, func(t *testing.T) {
			type contextKey struct{}
			ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), contextKey{}, "caller"), time.Hour)
			defer cancel()

			storage := &accessHashContextStorage{
				checkContext: func(got context.Context) {
					require.Equal(t, "caller", got.Value(contextKey{}))
					deadline, ok := got.Deadline()
					require.True(t, ok)
					want, _ := ctx.Deadline()
					require.Equal(t, want, deadline)
					require.Equal(t, ctx.Err(), got.Err())
				},
			}
			m := Options{Storage: storage}.Build(nil)
			self := getTestSelf()
			m.me.Store(self)

			const channelID, accessHash = int64(42), int64(123)
			require.NoError(t, storage.InmemoryStorage.Save(ctx,
				Key{Prefix: channelPrefix, ID: channelID}, Value{AccessHash: accessHash}))

			call := func() error {
				if method == "Set" {
					return m.SetChannelAccessHash(ctx, self.ID, channelID, accessHash)
				}
				hash, found, err := m.GetChannelAccessHash(ctx, self.ID, channelID)
				if err == nil {
					require.True(t, found)
					require.Equal(t, accessHash, hash)
				}
				return err
			}
			require.NoError(t, call())
			cancel()
			require.ErrorIs(t, call(), context.Canceled)
		})
	}
}
