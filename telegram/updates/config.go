package updates

import (
	"context"

	"github.com/gotd/log"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// API is the interface which contains
// Telegram RPC methods used by manager for internalState synchronization.
type API interface {
	UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error)
	UpdatesGetDifference(ctx context.Context, request *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error)
	UpdatesGetChannelDifference(ctx context.Context, request *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error)
}

// Config of the manager.
type Config struct {
	// Handler where updates will be passed.
	Handler telegram.UpdateHandler
	// Callback called if manager cannot
	// recover channel gap (optional).
	OnChannelTooLong func(channelID int64)
	// Callback called when the manager loses access to a channel, detected via
	// CHANNEL_PRIVATE on updates.getChannelDifference (e.g. the account was
	// kicked/banned or the channel was deleted). The channel is removed from
	// the update manager after this call (optional).
	OnChannelInaccessible func(channelID int64)
	// Callback called if manager cannot recover
	// common state gap, i.e. on updates.differenceTooLong (optional).
	OnTooLong func()
	// Callback called when the manager fails to load the locally stored user
	// state during Run with forget=false (no state found), so the state is
	// fetched from the server and a full resynchronization is performed.
	//
	// Useful to detect that updates missed while offline may need to be fetched
	// manually, e.g. via messages.getHistory (optional).
	OnLoadUserStateFailed func()
	// Callback called when the manager fails to load the locally stored state of
	// a channel during Run with forget=false (its access hash is missing), so
	// the channel is skipped.
	//
	// Useful to detect that the channel may need to be resynchronized manually
	// (optional).
	OnLoadChannelStateFailed func(channelID int64)
	// State storage.
	// In-mem used if not provided.
	Storage StateStorage
	// Channel access hash storage.
	// In-mem used if not provided.
	AccessHasher ChannelAccessHasher
	// User access hash storage.
	// In-mem used if not provided.
	UserAccessHasher UserAccessHasher
	// Logger (optional).
	Logger log.Logger
	// TracerProvider (optional).
	TracerProvider trace.TracerProvider
	// MaxChannelDifferenceConcurrency limits how many updates.getChannelDifference
	// requests may be in flight at once across all tracked channels. 0 (the
	// default) means unlimited: every channel recovers its gap independently, as
	// before. A positive value bounds the burst so an account in many active
	// channels does not exceed Telegram's per-account method rate limit.
	MaxChannelDifferenceConcurrency int
}

func (cfg *Config) setDefaults() {
	if cfg.Handler == nil {
		panic("Handler is nil")
	}
	if cfg.AccessHasher == nil {
		cfg.AccessHasher = newMemAccessHasher()
	}
	if cfg.UserAccessHasher == nil {
		cfg.UserAccessHasher = newMemUserAccessHasher()
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Nop
	}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = trace.NewNoopTracerProvider()
	}
	if cfg.Storage == nil {
		cfg.Storage = newMemStorage()
	}
	lg := log.For(cfg.Logger)
	if cfg.OnChannelTooLong == nil {
		cfg.OnChannelTooLong = func(channelID int64) {
			lg.Error(context.Background(), "Difference too long", log.Int64("channel_id", channelID))
		}
	}
	if cfg.OnChannelInaccessible == nil {
		cfg.OnChannelInaccessible = func(channelID int64) {
			lg.Info(context.Background(), "Channel is inaccessible, stopping updates",
				log.Int64("channel_id", channelID))
		}
	}
	if cfg.OnTooLong == nil {
		cfg.OnTooLong = func() {
			lg.Error(context.Background(), "Difference too long")
		}
	}
	if cfg.OnLoadUserStateFailed == nil {
		cfg.OnLoadUserStateFailed = func() {
			lg.Warn(context.Background(), "Failed to load user state, fetching from server")
		}
	}
	if cfg.OnLoadChannelStateFailed == nil {
		cfg.OnLoadChannelStateFailed = func(channelID int64) {
			lg.Warn(context.Background(), "Failed to load channel state, skipping channel",
				log.Int64("channel_id", channelID))
		}
	}
}
