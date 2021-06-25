package updates

import (
	"context"

	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

// RawClient is the interface which contains
// Telegram RPC methods used by the engine for state synchronization.
type RawClient interface {
	UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error)
	UpdatesGetDifference(ctx context.Context, request *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error)
	UpdatesGetChannelDifference(ctx context.Context, request *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error)
}

// Config of the engine.
type Config struct {
	RawClient    RawClient
	Handler      Handler
	Storage      Storage
	AccessHasher AccessHasher
	SelfID       int
	IsBot        bool
	Forget       bool
	Logger       *zap.Logger
}

func (cfg *Config) setDefaults() {
	if cfg.RawClient == nil {
		panic("raw client is nil")
	}
	if cfg.Handler == nil {
		panic("handler is nil")
	}
	if cfg.Storage == nil {
		cfg.Storage = NewMemStorage()
	}
	if cfg.AccessHasher == nil {
		cfg.AccessHasher = newMemAccessHasher()
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
}
