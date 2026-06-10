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
