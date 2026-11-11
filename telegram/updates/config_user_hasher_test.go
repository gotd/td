package updates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func TestConfigDefaultsUserAccessHasher(t *testing.T) {
	cfg := Config{Handler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		return nil
	})}
	cfg.setDefaults()
	require.NotNil(t, cfg.UserAccessHasher, "setDefaults must supply an in-mem UserAccessHasher")
}

func TestConfigPreservesUserAccessHasher(t *testing.T) {
	custom := newMemUserAccessHasher()
	cfg := Config{
		Handler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
			return nil
		}),
		UserAccessHasher: custom,
	}
	cfg.setDefaults()
	require.Same(t, custom, cfg.UserAccessHasher, "a provided UserAccessHasher must be preserved")
}

// PEER-5: a hasher that implements both interfaces is used for users too, so
// a persistent AccessHasher does not silently keep users in memory.
func TestConfigUserAccessHasherFromAccessHasher(t *testing.T) {
	both := NewFileStorage(t.TempDir() + "/updates.json")
	cfg := Config{
		Handler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
			return nil
		}),
		AccessHasher: both,
	}
	cfg.setDefaults()
	require.Same(t, UserAccessHasher(both), cfg.UserAccessHasher)

	// A channel-only hasher keeps the in-mem user default.
	cfg = Config{
		Handler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
			return nil
		}),
		AccessHasher: newMemAccessHasher(),
	}
	cfg.setDefaults()
	_, isMem := cfg.UserAccessHasher.(*memUserAccessHasher)
	require.True(t, isMem)
}
