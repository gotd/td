package updates

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// gateChannelDiffAPI is an API stub whose UpdatesGetChannelDifference blocks on
// release while tracking how many calls are in flight at once, so a test can
// observe the peak concurrency the updates manager allows.
type gateChannelDiffAPI struct {
	entered chan struct{} // one token per call that reached the RPC body
	release chan struct{} // calls block here until the test closes it

	mu      sync.Mutex
	current int
	peak    int
	total   int
}

func (a *gateChannelDiffAPI) UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error) {
	return &tg.UpdatesState{}, nil
}

func (a *gateChannelDiffAPI) UpdatesGetDifference(
	ctx context.Context, request *tg.UpdatesGetDifferenceRequest,
) (tg.UpdatesDifferenceClass, error) {
	return &tg.UpdatesDifferenceEmpty{}, nil
}

func (a *gateChannelDiffAPI) UpdatesGetChannelDifference(
	ctx context.Context, request *tg.UpdatesGetChannelDifferenceRequest,
) (tg.UpdatesChannelDifferenceClass, error) {
	a.mu.Lock()
	a.total++
	a.current++
	if a.current > a.peak {
		a.peak = a.current
	}
	a.mu.Unlock()

	a.entered <- struct{}{}
	<-a.release

	a.mu.Lock()
	a.current--
	a.mu.Unlock()
	return &tg.UpdatesChannelDifferenceEmpty{}, nil
}

func (a *gateChannelDiffAPI) snapshot() (current, peak, total int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.current, a.peak, a.total
}

// newGatedChannelState builds a channelState wired to the gate API and the
// given semaphore. Each channelState.Run issues exactly one getChannelDifference
// at startup (the gate API returns an empty, final difference).
func newGatedChannelState(t *testing.T, id int64, api API, sem chDiffSem) *channelState {
	t.Helper()
	return newChannelState(channelStateConfig{
		Out:        make(chan tracedUpdate, 10),
		ChannelID:  id,
		AccessHash: 1,
		RawClient:  api,
		Storage:    newMemStorage(),
		Handler: telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error {
			return nil
		}),
		OnChannelTooLong:      func(int64) {},
		OnChannelInaccessible: func(int64) {},
		RemoveChannel:         make(chan int64, 1),
		Logger:                logzap.New(zaptest.NewLogger(t)),
		Tracer:                noop.NewTracerProvider().Tracer(""),
		ChannelDiffSem:        sem,
	})
}

// TestNewThreadsChannelDiffConcurrency verifies Config.MaxChannelDifferenceConcurrency
// builds the manager's shared semaphore (and leaves it nil when unset).
func TestNewThreadsChannelDiffConcurrency(t *testing.T) {
	handler := telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil })

	m := New(Config{Handler: handler, MaxChannelDifferenceConcurrency: 3})
	require.NotNil(t, m.chDiffSem, "positive limit must build a semaphore")
	require.Equal(t, 3, cap(m.chDiffSem), "semaphore capacity must equal the limit")

	def := New(Config{Handler: handler})
	require.Nil(t, def.chDiffSem, "default (0) must leave the manager unlimited")
}

// TestNewChannelStateInheritsConcurrencyLimit verifies a channel tracked by a
// running manager at runtime (internalState.newChannelState, the path taken when
// an update arrives for a not-yet-tracked channel) shares the manager's
// getChannelDifference semaphore, so runtime joins are bounded by the same limit
// as the channels loaded at start.
func TestNewChannelStateInheritsConcurrencyLimit(t *testing.T) {
	ctx := context.Background()

	const selfID = 123
	storage := newMemStorage()
	require.NoError(t, storage.SetState(ctx, selfID, State{}))

	sem := newChDiffSem(4)
	s := newState(ctx, stateConfig{
		RawClient: &gateChannelDiffAPI{},
		Logger:    logzap.New(zaptest.NewLogger(t)),
		Tracer:    noop.NewTracerProvider().Tracer(""),
		Handler: telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error {
			return nil
		}),
		OnChannelTooLong: func(int64) {},
		Storage:          storage,
		Hasher:           newMemAccessHasher(),
		SelfID:           selfID,
		DiffLimit:        diffLimitUser,
		WorkGroup:        &errgroup.Group{},
		ChannelDiffSem:   sem,
	})

	ch := s.newChannelState(4242, 1, 0)
	require.Equal(t, sem, ch.chDiffSem,
		"a runtime-added channel must share the manager's getChannelDifference semaphore")
}

// TestChannelDifferenceConcurrencyLimited verifies the shared semaphore caps the
// number of concurrent getChannelDifference calls across channels.
func TestChannelDifferenceConcurrencyLimited(t *testing.T) {
	const (
		limit    = 2
		channels = 6
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	api := &gateChannelDiffAPI{
		entered: make(chan struct{}, channels),
		release: make(chan struct{}),
	}
	sem := newChDiffSem(limit)

	var wg sync.WaitGroup
	for i := range channels {
		st := newGatedChannelState(t, int64(1000+i), api, sem)
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = st.Run(ctx)
		}()
	}

	// Exactly `limit` calls may reach the RPC; collect them.
	for range limit {
		<-api.entered
	}
	// No further call may start while the slots are held.
	select {
	case <-api.entered:
		t.Fatal("getChannelDifference exceeded the concurrency limit")
	case <-time.After(100 * time.Millisecond):
	}
	current, _, _ := api.snapshot()
	require.Equal(t, limit, current, "exactly `limit` calls must be in flight")

	// Release the held calls; the remaining channels then proceed, still capped.
	close(api.release)
	for range channels - limit {
		<-api.entered
	}

	cancel()
	wg.Wait()

	_, peak, total := api.snapshot()
	require.LessOrEqual(t, peak, limit, "peak concurrency must never exceed the limit")
	require.Equal(t, channels, total, "every channel must eventually fetch its difference")
}

// TestChannelDifferenceUnlimitedByDefault verifies a zero limit (nil semaphore)
// keeps today's behavior: every channel recovers concurrently, uncapped.
func TestChannelDifferenceUnlimitedByDefault(t *testing.T) {
	const channels = 6

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	api := &gateChannelDiffAPI{
		entered: make(chan struct{}, channels),
		release: make(chan struct{}),
	}
	sem := newChDiffSem(0) // unlimited

	var wg sync.WaitGroup
	for i := range channels {
		st := newGatedChannelState(t, int64(2000+i), api, sem)
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = st.Run(ctx)
		}()
	}

	// All channels must be able to enter the RPC at once; if a cap were wrongly
	// applied this would block and the test would time out.
	for range channels {
		<-api.entered
	}
	current, _, _ := api.snapshot()
	require.Equal(t, channels, current, "all channels must run concurrently when unlimited")

	close(api.release)
	cancel()
	wg.Wait()

	_, peak, _ := api.snapshot()
	require.Equal(t, channels, peak, "peak concurrency must reach all channels when unlimited")
}

// TestChannelDifferenceAcquireRespectsContext verifies acquire aborts on context
// cancellation without issuing the RPC.
func TestChannelDifferenceAcquireRespectsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	api := &gateChannelDiffAPI{
		entered: make(chan struct{}, 2),
		release: make(chan struct{}),
	}
	sem := newChDiffSem(1)

	var wg sync.WaitGroup

	// Channel A grabs the only slot and blocks inside the RPC.
	stA := newGatedChannelState(t, 3001, api, sem)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = stA.Run(ctx)
	}()
	<-api.entered // A is in flight, holding the slot.

	// Channel B blocks on acquire and must never reach the RPC.
	stB := newGatedChannelState(t, 3002, api, sem)
	bDone := make(chan struct{})
	go func() {
		_ = stB.Run(ctx)
		close(bDone)
	}()
	select {
	case <-api.entered:
		t.Fatal("B entered the RPC despite a full semaphore")
	case <-time.After(100 * time.Millisecond):
	}

	cancel() // B's acquire must return ctx.Err and B.Run must finish.
	<-bDone

	_, _, total := api.snapshot()
	require.Equal(t, 1, total, "B must not call getChannelDifference after ctx cancel")

	close(api.release) // let A unwind.
	wg.Wait()
}
