package updates

import (
	"context"

	"github.com/gotd/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/td/clock"
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
	// StopOnApplyError makes Run return on Handler or StateStorage errors
	// while applying updates, instead of logging them and continuing
	// (default).
	//
	// When enabled, a failed Handler call is not followed by the checkpoint
	// write of its apply path, so the pts, qts, seq or channel pts is not
	// advanced in memory or in Storage; a failed StateStorage write does not
	// advance the corresponding in-memory state; a difference slice is not
	// continued. Errors of RPC calls keep their usual handling.
	//
	// This is an error policy, not a transaction: completed steps and
	// concurrent work (e.g. other channels) are not rolled back, a Storage
	// may commit before returning an error, and a handler may have partially
	// applied the failed batch. Handlers must be idempotent, as updates
	// after the last successful checkpoint are delivered again. After
	// resolving the failure, restart with a new Manager and Forget=false to
	// resume from the stored state.
	//
	// Updates outside any sequence that a channel difference forwards to the
	// common state are handed off before the channel pts is stored, so a
	// failure to handle them does not hold that pts back.
	StopOnApplyError bool
	// Callback called if manager cannot
	// recover channel gap (optional).
	//
	// See also OnChannelGap, which provides the gap range.
	OnChannelTooLong func(channelID int64)
	// Callback called if manager cannot recover channel gap, i.e. on
	// updates.channelDifferenceTooLong, with the gap range (optional).
	//
	// The latest channel messages returned by the server with
	// channelDifferenceTooLong are passed in gap.Messages, not to Handler:
	// they may include messages delivered before the gap. Messages between
	// the last known message and gap.Messages should be fetched manually,
	// e.g. via messages.getHistory.
	//
	// Called after OnChannelTooLong.
	OnChannelGap func(gap ChannelGap)
	// Callback called when the manager loses access to a channel, detected via
	// CHANNEL_PRIVATE or CHANNEL_INVALID on updates.getChannelDifference (e.g. the account was
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
	// Useful to detect that the channel may need to be resynchronized manually,
	// e.g. by calling Manager.TrackChannel with its access hash, which resumes
	// from the stored pts (optional).
	OnLoadChannelStateFailed func(channelID int64)
	// State storage.
	// In-mem used if not provided.
	Storage StateStorage
	// Channel access hash storage. In-mem used if not provided.
	//
	// The manager needs the access hash of a channel to track it. Hashes are
	// learned from the entities of updates and differences; for a channel
	// the manager has not seen there yet (e.g. known only from dialogs), its
	// first update costs an extra updates.getDifference to find the hash.
	// Pass a hasher that also knows the peers the rest of the app has seen,
	// such as *peers.Manager, or a persistent one (FileStorage, the contrib
	// storages), to avoid that and to keep channels across restarts.
	AccessHasher ChannelAccessHasher
	// User access hash storage. If not provided, AccessHasher is used when
	// it implements UserAccessHasher (e.g. *peers.Manager, FileStorage),
	// otherwise an in-mem one.
	UserAccessHasher UserAccessHasher
	// Logger (optional).
	Logger log.Logger
	// TracerProvider (optional).
	TracerProvider trace.TracerProvider
	// MeterProvider for metrics of the manager (optional). No metrics by
	// default. See the package documentation for the list.
	MeterProvider metric.MeterProvider
	// MaxChannelDifferenceConcurrency limits how many updates.getChannelDifference
	// requests may be in flight at once across all tracked channels. Non-positive
	// values (<= 0, default 0) mean unlimited: every channel recovers its gap as
	// before. A positive value bounds the burst so an account in many active
	// channels does not exceed Telegram's per-account method rate limit.
	MaxChannelDifferenceConcurrency int
	// BeforeRun is what to do with updates received before Run, e.g. while
	// the client connects and authorizes. Default is BeforeRunPass.
	//
	// Dropped updates (BeforeRunDrop, or over BeforeRunBufferSize) of the
	// common or a tracked channel sequence are recovered via getDifference
	// once the next update reveals the gap.
	BeforeRun BeforeRunMode
	// BeforeRunBufferSize limits how many update batches are kept with
	// BeforeRunBuffer. Non-positive means 100.
	BeforeRunBufferSize int

	// Clock is used for retries of failed differences. Defaults to
	// clock.System.
	Clock clock.Clock
}

func (cfg *Config) setDefaults() {
	if cfg.Handler == nil {
		panic("Handler is nil")
	}
	if cfg.AccessHasher == nil {
		cfg.AccessHasher = newMemAccessHasher()
	}
	if cfg.Clock == nil {
		cfg.Clock = clock.System
	}
	if cfg.UserAccessHasher == nil {
		if h, ok := cfg.AccessHasher.(UserAccessHasher); ok {
			// One hasher for both, e.g. a persistent one: users kept in
			// memory would be lost on restart while channels are kept.
			cfg.UserAccessHasher = h
		} else {
			cfg.UserAccessHasher = newMemUserAccessHasher()
		}
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Nop
	}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = trace.NewNoopTracerProvider()
	}
	if cfg.MeterProvider == nil {
		cfg.MeterProvider = noop.NewMeterProvider()
	}
	if cfg.Storage == nil {
		cfg.Storage = newMemStorage()
	}
	if cfg.BeforeRunBufferSize <= 0 {
		cfg.BeforeRunBufferSize = defaultBeforeRunBufferSize
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
