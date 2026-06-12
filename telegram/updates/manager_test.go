package updates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func newTestManager(t *testing.T, cfg Config) *Manager {
	cfg.Handler = telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil })
	cfg.Logger = logzap.New(zaptest.NewLogger(t))
	return New(cfg)
}

func TestManager_loadState(t *testing.T) {
	ctx := context.Background()
	const userID = 1

	t.Run("Found", func(t *testing.T) {
		a := require.New(t)
		called := false
		storage := newMemStorage()
		a.NoError(storage.SetState(ctx, userID, State{Pts: 42}))

		m := newTestManager(t, Config{
			Storage:               storage,
			OnLoadUserStateFailed: func() { called = true },
		})

		state, err := m.loadState(ctx, &diffAPI{}, userID, false)
		a.NoError(err)
		a.Equal(42, state.Pts)
		a.False(called, "callback must not fire when state is found")
	})

	t.Run("NotFound", func(t *testing.T) {
		a := require.New(t)
		called := false

		m := newTestManager(t, Config{
			Storage:               newMemStorage(),
			OnLoadUserStateFailed: func() { called = true },
		})

		// No stored state: falls back to remote (forget) and reports failure.
		_, err := m.loadState(ctx, &diffAPI{}, userID, false)
		a.NoError(err)
		a.True(called, "callback must fire when state is not found")
	})

	t.Run("Forget", func(t *testing.T) {
		a := require.New(t)
		called := false

		m := newTestManager(t, Config{
			Storage:               newMemStorage(),
			OnLoadUserStateFailed: func() { called = true },
		})

		// Explicit forget must not report a load failure.
		_, err := m.loadState(ctx, &diffAPI{}, userID, true)
		a.NoError(err)
		a.False(called, "callback must not fire on explicit forget")
	})
}

func TestManager_loadChannels(t *testing.T) {
	ctx := context.Background()
	const userID = 1

	t.Run("MissingAccessHash", func(t *testing.T) {
		a := require.New(t)
		var failed []int64

		storage := newMemStorage()
		a.NoError(storage.SetState(ctx, userID, State{}))
		a.NoError(storage.SetChannelPts(ctx, userID, 10, 1))
		a.NoError(storage.SetChannelPts(ctx, userID, 20, 2))

		hasher := newMemAccessHasher()
		// Only channel 10 has a known access hash.
		a.NoError(hasher.SetChannelAccessHash(ctx, userID, 10, 0xabc))

		m := newTestManager(t, Config{
			Storage:      storage,
			AccessHasher: hasher,
			OnLoadChannelStateFailed: func(channelID int64) {
				failed = append(failed, channelID)
			},
		})

		channels, err := m.loadChannels(ctx, userID)
		a.NoError(err)
		a.Len(channels, 1)
		a.Contains(channels, int64(10))
		a.Equal(int64(0xabc), channels[10].AccessHash)
		a.Equal([]int64{20}, failed)
	})

	t.Run("AllPresent", func(t *testing.T) {
		a := require.New(t)
		called := false

		storage := newMemStorage()
		a.NoError(storage.SetState(ctx, userID, State{}))
		a.NoError(storage.SetChannelPts(ctx, userID, 10, 1))

		hasher := newMemAccessHasher()
		a.NoError(hasher.SetChannelAccessHash(ctx, userID, 10, 0xabc))

		m := newTestManager(t, Config{
			Storage:                  storage,
			AccessHasher:             hasher,
			OnLoadChannelStateFailed: func(int64) { called = true },
		})

		channels, err := m.loadChannels(ctx, userID)
		a.NoError(err)
		a.Len(channels, 1)
		a.False(called, "callback must not fire when all access hashes are present")
	})
}
