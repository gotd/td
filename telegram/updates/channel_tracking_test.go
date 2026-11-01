package updates

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type trackingAPI struct {
	diffAPI
	common  func(context.Context, *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error)
	channel func(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error)
}

func (a *trackingAPI) UpdatesGetDifference(ctx context.Context, r *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error) {
	if a.common != nil {
		return a.common(ctx, r)
	}
	return a.diffAPI.UpdatesGetDifference(ctx, r)
}

func (a *trackingAPI) UpdatesGetChannelDifference(ctx context.Context, r *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
	return a.channel(ctx, r)
}

type trackingStorage struct {
	StateStorage
	scan  func(context.Context, int64, func(context.Context, int64, int) error) error
	read  func(context.Context, int64, int64) (int, bool, error)
	write func(context.Context, int64, int64, int) error
}

func (s *trackingStorage) SetChannelPts(ctx context.Context, userID, channelID int64, pts int) error {
	if s.write != nil {
		return s.write(ctx, userID, channelID, pts)
	}
	return s.StateStorage.SetChannelPts(ctx, userID, channelID, pts)
}

func (s *trackingStorage) GetChannelPts(ctx context.Context, userID, channelID int64) (pts int, found bool, err error) {
	if s.read != nil {
		return s.read(ctx, userID, channelID)
	}
	return s.StateStorage.GetChannelPts(ctx, userID, channelID)
}

func (s *trackingStorage) ForEachChannels(ctx context.Context, id int64, f func(context.Context, int64, int) error) error {
	if s.scan != nil {
		return s.scan(ctx, id, f)
	}
	return s.StateStorage.ForEachChannels(ctx, id, f)
}

type trackingHasher struct {
	ChannelAccessHasher
	reads  []int64
	lookup func(context.Context, int64, int64) (int64, bool, error)
}

func (h *trackingHasher) GetChannelAccessHash(ctx context.Context, userID, channelID int64) (accessHash int64, found bool, err error) {
	h.reads = append(h.reads, channelID)
	if h.lookup != nil {
		return h.lookup(ctx, userID, channelID)
	}
	return h.ChannelAccessHasher.GetChannelAccessHash(ctx, userID, channelID)
}

func trackingFixture(t *testing.T) (*memStorage, *memAccessHasher) {
	t.Helper()
	ctx := context.Background()
	storage, hasher := newMemStorage(), newMemAccessHasher()
	require.NoError(t, storage.SetState(ctx, 1, State{}))
	require.NoError(t, storage.SetChannelPts(ctx, 1, 41, 10))
	require.NoError(t, hasher.SetChannelAccessHash(ctx, 1, 41, 123))
	return storage, hasher
}

func runTrackingManager(t *testing.T, m *Manager, api API, opt AuthOptions) context.CancelFunc {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- m.Run(ctx, api, 1, opt) }()
	synctest.Wait()
	t.Cleanup(func() {
		cancel()
		require.ErrorIs(t, <-done, context.Canceled)
	})
	return cancel
}

func TestChannelTrackingRefresh(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		storage, hashes := trackingFixture(t)
		// An excluded channel must not even look up a missing hash at boot.
		require.NoError(t, storage.SetChannelPts(context.Background(), 1, 42, 20))
		hasher := &trackingHasher{ChannelAccessHasher: hashes}
		var selected atomic.Bool
		var requests []int
		var failed []int64
		api := &trackingAPI{
			diffAPI: diffAPI{diffs: []tg.UpdatesDifferenceClass{&tg.UpdatesDifferenceEmpty{}}},
			channel: func(ctx context.Context, r *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
				require.Equal(t, int64(41), r.Channel.(*tg.InputChannel).ChannelID)
				requests = append(requests, r.Pts)
				return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: r.Pts}, nil
			},
		}
		var handled int
		var m *Manager
		m = New(Config{
			Storage: storage, AccessHasher: hasher,
			TrackChannel:             func(id int64) bool { return id == 41 && selected.Load() },
			OnLoadChannelStateFailed: func(id int64) { failed = append(failed, id) },
			Handler: telegram.UpdateHandlerFunc(func(ctx context.Context, _ tg.UpdatesClass) error {
				handled++
				// No acknowledgment round trip to the common owner.
				return m.RefreshChannelTracking(ctx)
			}),
		})
		runTrackingManager(t, m, api, AuthOptions{})
		require.Empty(t, hasher.reads)
		require.Empty(t, requests)
		require.Empty(t, failed)

		selected.Store(true)
		callCtx, cancelCall := context.WithCancel(context.Background())
		require.NoError(t, m.RefreshChannelTracking(callCtx))
		cancelCall() // Workers belong to Run, not this request.
		synctest.Wait()
		require.Equal(t, []int{10}, requests, "a quiet saved channel starts without a push")

		selected.Store(false)
		require.NoError(t, m.RefreshChannelTracking(context.Background()))
		synctest.Wait()
		require.Empty(t, m.state.channels)
		require.Empty(t, m.state.retiring, "completed workers must be released without another refresh")
		require.NoError(t, m.Handle(context.Background(), &tg.UpdateShort{Update: &tg.UpdateUserStatus{UserID: 7}}))
		synctest.Wait()
		require.Equal(t, 1, handled, "common updates still reach the handler")
		require.Equal(t, []int{10}, requests)

		selected.Store(true)
		require.NoError(t, m.RefreshChannelTracking(context.Background()))
		synctest.Wait()
		require.Equal(t, []int{10, 10}, requests)
	})
}

func TestChannelTrackingOnStartAndDefault(t *testing.T) {
	for _, selection := range []bool{false, true} {
		name := "Default"
		if selection {
			name = "OnStart"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				storage, hasher := trackingFixture(t)
				var selected atomic.Bool
				var tooLong []int64
				cfg := Config{Storage: storage, AccessHasher: hasher,
					Handler:          telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil }),
					OnChannelTooLong: func(id int64) { tooLong = append(tooLong, id) },
				}
				if selection {
					cfg.TrackChannel = func(int64) bool { return selected.Load() }
				}
				m := New(cfg)
				var requests int
				api := &trackingAPI{
					diffAPI: diffAPI{diffs: []tg.UpdatesDifferenceClass{&tg.UpdatesDifferenceEmpty{}}},
					channel: func(ctx context.Context, r *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
						requests++
						if selection {
							// A saved quiet channel can be too far behind when enabled.
							dialog := &tg.Dialog{}
							dialog.SetPts(1000)
							return &tg.UpdatesChannelDifferenceTooLong{Final: true, Dialog: dialog}, nil
						}
						return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: r.Pts}, nil
					},
				}
				runTrackingManager(t, m, api, AuthOptions{OnStart: func(ctx context.Context) {
					selected.Store(true)
					require.NoError(t, m.RefreshChannelTracking(ctx))
					require.NoError(t, m.RefreshChannelTracking(ctx)) // Coalesce before the owner starts.
				}})
				require.Equal(t, 1, requests)
				if !selection {
					require.Nil(t, m.state.trackingQueue)
					require.Nil(t, m.state.retiring)
				} else {
					require.Equal(t, []int64{41}, tooLong)
					pts, found, err := storage.GetChannelPts(context.Background(), 1, 41)
					require.NoError(t, err)
					require.True(t, found)
					require.Equal(t, 1000, pts, "TooLong uses the server checkpoint and reports the gap")
				}
			})
		})
	}
}

func TestChannelTrackingScanFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		storage, hasher := trackingFixture(t)
		wrapped := &trackingStorage{StateStorage: storage}
		m := newTestManager(t, Config{Storage: wrapped, AccessHasher: hasher, TrackChannel: func(int64) bool { return true }})
		api := &trackingAPI{
			diffAPI: diffAPI{diffs: []tg.UpdatesDifferenceClass{&tg.UpdatesDifferenceEmpty{}}},
			channel: func(ctx context.Context, r *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
				return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: r.Pts}, nil
			},
		}
		require.Error(t, m.RefreshChannelTracking(context.Background()), "before Run")
		cancel := runTrackingManager(t, m, api, AuthOptions{})
		want := errors.New("scan failed after a row")
		wrapped.scan = func(ctx context.Context, _ int64, f func(context.Context, int64, int) error) error {
			require.NoError(t, f(ctx, 99, 1))
			return want
		}
		require.ErrorIs(t, m.RefreshChannelTracking(context.Background()), want)
		require.Empty(t, m.state.trackingQueue, "never publish a partial scan")
		require.Contains(t, m.state.channels, int64(41))
		ctx, cancelCall := context.WithCancel(context.Background())
		cancelCall()
		require.ErrorIs(t, m.RefreshChannelTracking(ctx), context.Canceled)
		cancel()
		synctest.Wait()
		require.Error(t, m.RefreshChannelTracking(context.Background()), "after Run")
	})
}

func TestChannelTrackingRefreshReadFailure(t *testing.T) {
	for _, phase := range []string{"enumeration", "before hash", "after hash", "hash error"} {
		t.Run(phase, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				storage, hashes := trackingFixture(t)
				wrapped := &trackingStorage{StateStorage: storage}
				hasher := &trackingHasher{ChannelAccessHasher: hashes}
				var selected atomic.Bool
				m := newTestManager(t, Config{Storage: wrapped, AccessHasher: hasher, TrackChannel: func(int64) bool { return selected.Load() }})
				api := &trackingAPI{diffAPI: diffAPI{diffs: []tg.UpdatesDifferenceClass{&tg.UpdatesDifferenceEmpty{}}},
					channel: func(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
						t.Error("failed refresh must not start recovery")
						return nil, errors.New("unexpected RPC")
					},
				}
				runTrackingManager(t, m, api, AuthOptions{})
				selected.Store(true)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				want := context.Canceled
				if phase == "hash error" {
					want = errors.New("hash storage unavailable")
				}
				wrapped.scan = func(ctx context.Context, _ int64, f func(context.Context, int64, int) error) error {
					if phase == "enumeration" {
						cancel()
					}
					err := f(ctx, 41, 10)
					if phase == "before hash" {
						cancel()
					}
					return err
				}
				hasher.lookup = func(ctx context.Context, userID, channelID int64) (int64, bool, error) {
					if phase == "hash error" {
						return 0, false, want
					}
					if phase == "after hash" {
						cancel()
					}
					return hashes.GetChannelAccessHash(ctx, userID, channelID)
				}
				require.ErrorIs(t, m.RefreshChannelTracking(ctx), want)
				require.Empty(t, m.state.trackingQueue, "a failed or canceled read must not publish candidates")
				synctest.Wait()
				require.Empty(t, m.state.channels)
			})
		})
	}
}

func TestChannelTrackingSelectionChangesDuringStartup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		storage, hashes := trackingFixture(t)
		var selected atomic.Bool
		selected.Store(true)
		hasher := &trackingHasher{ChannelAccessHasher: hashes}
		hasher.lookup = func(ctx context.Context, userID, channelID int64) (int64, bool, error) {
			selected.Store(false) // Policy changes while startup waits for storage.
			return hashes.GetChannelAccessHash(ctx, userID, channelID)
		}
		m := newTestManager(t, Config{Storage: storage, AccessHasher: hasher, TrackChannel: func(int64) bool { return selected.Load() }})
		api := &trackingAPI{diffAPI: diffAPI{diffs: []tg.UpdatesDifferenceClass{&tg.UpdatesDifferenceEmpty{}}},
			channel: func(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
				t.Error("startup must recheck selection before starting a worker")
				return nil, errors.New("unexpected RPC")
			},
		}
		runTrackingManager(t, m, api, AuthOptions{})
		require.Empty(t, m.state.channels)
	})
}

func TestChannelTrackingDrainsBeforeReenable(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		storage, hasher := trackingFixture(t)
		var selected atomic.Bool
		selected.Store(true)
		release := make(chan struct{})
		var requests []int
		var callbackContext context.Context
		var common int
		api := &trackingAPI{
			diffAPI: diffAPI{diffs: []tg.UpdatesDifferenceClass{&tg.UpdatesDifferenceEmpty{}}},
			channel: func(ctx context.Context, r *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
				requests = append(requests, r.Pts)
				if len(requests) == 1 {
					return &tg.UpdatesChannelDifference{Final: true, Pts: 11,
						NewMessages: []tg.MessageClass{&tg.Message{ID: 1, PeerID: &tg.PeerChannel{ChannelID: 41}}},
					}, nil
				}
				return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: r.Pts}, nil
			},
		}
		m := New(Config{
			Storage: storage, AccessHasher: hasher,
			TrackChannel: func(int64) bool { return selected.Load() },
			Handler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
				if _, channel := u.(*tg.Updates).Updates[0].(*tg.UpdateNewChannelMessage); !channel {
					common++
					return nil
				}
				callbackContext = ctx
				<-release // Represents an already-admitted durable transaction.
				return ctx.Err()
			}),
		})
		runTrackingManager(t, m, api, AuthOptions{})
		require.NotNil(t, callbackContext)
		first := m.state.channels[41]

		// A -> B -> C, with B canceled before A finishes.
		for _, enabled := range []bool{false, true, false, true} {
			selected.Store(enabled)
			require.NoError(t, m.RefreshChannelTracking(context.Background()))
			synctest.Wait()
		}
		require.Equal(t, []int{10}, requests, "neither replacement may overtake the old transaction")
		require.NoError(t, callbackContext.Err(), "deselection must not cancel admitted persistence")
		require.NoError(t, m.Handle(context.Background(), &tg.UpdateShort{Update: &tg.UpdateUserStatus{UserID: 7}}))
		synctest.Wait()
		require.Equal(t, 1, common, "the common owner must not wait for retirement")

		close(release)
		synctest.Wait()
		require.Equal(t, []int{10, 11}, requests, "resume must reread the late committed PTS")
		pts, found, err := storage.GetChannelPts(context.Background(), 1, 41)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, 11, pts)
		require.Empty(t, m.state.retiring)
		replacement := m.state.channels[41]
		require.NotSame(t, first, replacement)
		m.state.removeChannel <- first // Delayed inaccessible notification.
		synctest.Wait()
		require.Same(t, replacement, m.state.channels[41])
	})
}

func TestChannelTrackingOlderScanCannotReplaceNewer(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		storage, hasher := trackingFixture(t)
		require.NoError(t, storage.SetChannelPts(context.Background(), 1, 42, 20))
		require.NoError(t, hasher.SetChannelAccessHash(context.Background(), 1, 42, 456))
		wrapped := &trackingStorage{StateStorage: storage}
		var selected atomic.Bool
		var requests []int64
		m := newTestManager(t, Config{Storage: wrapped, AccessHasher: hasher,
			TrackChannel: func(int64) bool { return selected.Load() },
		})
		api := &trackingAPI{
			diffAPI: diffAPI{diffs: []tg.UpdatesDifferenceClass{&tg.UpdatesDifferenceEmpty{}}},
			channel: func(ctx context.Context, r *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
				requests = append(requests, r.Channel.(*tg.InputChannel).ChannelID)
				return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: r.Pts}, nil
			},
		}
		runTrackingManager(t, m, api, AuthOptions{})
		selected.Store(true)
		release := make(chan struct{})
		var scans atomic.Int32
		wrapped.scan = func(ctx context.Context, _ int64, f func(context.Context, int64, int) error) error {
			if scans.Add(1) == 1 {
				if err := f(ctx, 41, 10); err != nil {
					return err
				}
				<-release
				return nil
			}
			return f(ctx, 42, 20)
		}
		older := make(chan error, 1)
		go func() { older <- m.RefreshChannelTracking(context.Background()) }()
		synctest.Wait()
		require.NoError(t, m.RefreshChannelTracking(context.Background()))
		synctest.Wait() // The newer snapshot has already been consumed.
		require.Equal(t, []int64{42}, requests)
		close(release)
		require.NoError(t, <-older)
		synctest.Wait()
		require.Equal(t, []int64{42}, requests, "the older scan must remain superseded after consumption")
	})
}

func TestChannelTrackingScanCannotCrossRunLifetime(t *testing.T) {
	for _, reset := range []bool{false, true} {
		name := "Canceled"
		if reset {
			name = "Reset"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				storage, hasher := trackingFixture(t)
				wrapped := &trackingStorage{StateStorage: storage}
				var selected atomic.Bool
				m := newTestManager(t, Config{Storage: wrapped, AccessHasher: hasher,
					TrackChannel: func(int64) bool { return selected.Load() },
				})
				api := &trackingAPI{
					diffAPI: diffAPI{diffs: []tg.UpdatesDifferenceClass{&tg.UpdatesDifferenceEmpty{}}},
					channel: func(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
						t.Error("a stale refresh must not start a channel")
						return nil, errors.New("unexpected RPC")
					},
				}
				cancel := runTrackingManager(t, m, api, AuthOptions{})
				state := m.state
				selected.Store(true)
				release := make(chan struct{})
				wrapped.scan = func(ctx context.Context, _ int64, f func(context.Context, int64, int) error) error {
					<-release
					return f(ctx, 41, 10)
				}
				result := make(chan error, 1)
				go func() { result <- m.RefreshChannelTracking(context.Background()) }()
				synctest.Wait()
				if reset {
					m.Reset()
				} else {
					cancel()
				}
				close(release)
				require.Error(t, <-result)
				require.Empty(t, state.trackingQueue)
			})
		})
	}
}

func TestChannelTrackingRechecksPendingWork(t *testing.T) {
	for _, waitingForSlot := range []bool{true, false} {
		name := "QueuedUpdate"
		if waitingForSlot {
			name = "RPCSlot"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var selected atomic.Bool
				selected.Store(true)
				sem := newChDiffSem(1)
				if waitingForSlot {
					require.NoError(t, sem.acquire(context.Background()))
				}
				api := &trackingAPI{channel: func(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
					if waitingForSlot {
						t.Error("deselected channel must not issue an RPC after waiting for a slot")
					}
					return &tg.UpdatesChannelDifferenceEmpty{Final: true}, nil
				}}
				state := newGatedChannelState(t, 41, api, sem)
				state.trackChannel = func(int64) bool { return selected.Load() }
				done := make(chan error, 1)
				go func() { done <- state.Run(context.Background()) }()
				synctest.Wait()
				selected.Store(false)
				if waitingForSlot {
					sem.release()
					require.NoError(t, <-done)
				} else {
					// Work admitted before policy publication can arrive while
					// the worker is already waiting on its queue.
					state.updates <- channelUpdate{affected: true, pts: 1, ptsCount: 1}
					require.ErrorIs(t, <-done, errChannelDeselected)
					require.Zero(t, state.pts.State(), "queued work must not advance PTS after deselection")
				}
				require.True(t, state.getDifferenceLogger(context.Background(), "idle-recovery"),
					"recovery must recognize deselection as a clean stop")
			})
		})
	}
}

func TestChannelTrackingReenabledBeforeCompletion(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		storage, hasher := trackingFixture(t)
		var selected atomic.Bool
		selected.Store(true)
		rpcRelease, commonRelease := make(chan struct{}), make(chan struct{})
		var requests int
		api := &trackingAPI{
			common: func(ctx context.Context, _ *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error) {
				<-commonRelease
				return &tg.UpdatesDifferenceEmpty{}, nil
			},
			channel: func(ctx context.Context, r *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
				requests++
				if requests == 1 {
					<-rpcRelease
				}
				return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: r.Pts}, nil
			},
		}
		m := newTestManager(t, Config{Storage: storage, AccessHasher: hasher, TrackChannel: func(int64) bool { return selected.Load() }})
		runTrackingManager(t, m, api, AuthOptions{})
		selected.Store(false)
		close(rpcRelease)
		synctest.Wait() // The worker stopped, but its owner has not consumed completion.
		selected.Store(true)
		require.NoError(t, m.RefreshChannelTracking(context.Background()))
		close(commonRelease)
		synctest.Wait()
		require.Equal(t, 2, requests, "reenabling must not leave a completed registration in place")
		require.Len(t, m.state.channels, 1)
	})
}

func TestChannelTrackingLiveAdmission(t *testing.T) {
	for _, failure := range []string{"None", "InitialRead", "InitialWrite"} {
		t.Run(failure, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				storage, hashes := trackingFixture(t)
				wrapped := &trackingStorage{StateStorage: storage}
				hasher := &trackingHasher{ChannelAccessHasher: hashes}
				var selected atomic.Bool
				var requests, handled int
				api := &trackingAPI{
					diffAPI: diffAPI{diffs: []tg.UpdatesDifferenceClass{&tg.UpdatesDifferenceEmpty{}}},
					channel: func(ctx context.Context, r *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
						requests++
						require.Equal(t, 20, r.Pts, "first live update seeds only its preceding checkpoint")
						return &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: r.Pts}, nil
					},
				}
				m := New(Config{Storage: wrapped, AccessHasher: hasher, TrackChannel: func(int64) bool { return selected.Load() },
					Handler: telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { handled++; return nil }),
				})
				runTrackingManager(t, m, api, AuthOptions{})
				require.NoError(t, hashes.SetChannelAccessHash(context.Background(), 1, 43, 123))
				fault := errors.New("checkpoint unavailable")
				switch failure {
				case "InitialRead":
					wrapped.read = func(context.Context, int64, int64) (int, bool, error) { return 0, false, fault }
				case "InitialWrite":
					wrapped.write = func(ctx context.Context, userID, channelID int64, pts int) error {
						if requests == 0 {
							return fault
						}
						return storage.SetChannelPts(ctx, userID, channelID, pts)
					}
				}
				push := func() {
					require.NoError(t, m.Handle(context.Background(), &tg.UpdateShort{Update: &tg.UpdateDeleteChannelMessages{ChannelID: 43, Messages: []int{1}, Pts: 21, PtsCount: 1}}))
					synctest.Wait()
				}
				push()
				require.Empty(t, hasher.reads, "excluded pushes must not resolve hashes or initialize checkpoints")
				require.Empty(t, m.state.channels)
				selected.Store(true)
				push()
				require.Equal(t, 1, requests)
				require.Equal(t, 1, handled)
				require.NoError(t, m.HandleAffected(context.Background(), 43, 22, 1))
				synctest.Wait()
				selected.Store(false)
				push()
				require.NoError(t, m.HandleAffected(context.Background(), 43, 23, 1))
				require.NoError(t, m.Handle(context.Background(), &tg.UpdateShort{Update: &tg.UpdateChannelTooLong{ChannelID: 43}}))
				synctest.Wait()
				pts, found, err := storage.GetChannelPts(context.Background(), 1, 43)
				require.NoError(t, err)
				require.True(t, found)
				require.Equal(t, 22, pts, "excluded updates and affected results cannot advance PTS")
				require.Equal(t, 1, requests)
				require.Equal(t, 1, handled)
			})
		})
	}
}

func TestChannelTrackingMissingResumeCheckpoint(t *testing.T) {
	for _, failure := range []error{nil, errors.New("checkpoint read failed")} {
		name := "Missing"
		if failure != nil {
			name = "ReadError"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				storage, hasher := trackingFixture(t)
				wrapped := &trackingStorage{StateStorage: storage}
				var selected atomic.Bool
				var failed []int64
				m := newTestManager(t, Config{Storage: wrapped, AccessHasher: hasher, TrackChannel: func(int64) bool { return selected.Load() },
					OnLoadChannelStateFailed: func(id int64) { failed = append(failed, id) },
				})
				api := &trackingAPI{diffAPI: diffAPI{diffs: []tg.UpdatesDifferenceClass{&tg.UpdatesDifferenceEmpty{}}},
					channel: func(context.Context, *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error) {
						t.Error("missing resume state must not bootstrap to latest")
						return nil, errors.New("unexpected RPC")
					},
				}
				runTrackingManager(t, m, api, AuthOptions{})
				wrapped.read = func(context.Context, int64, int64) (int, bool, error) { return 0, false, failure }
				selected.Store(true)
				require.NoError(t, m.RefreshChannelTracking(context.Background()))
				synctest.Wait()
				require.Equal(t, []int64{41}, failed)
				require.Empty(t, m.state.channels)
			})
		})
	}
}

func BenchmarkChannelTrackingLoadChannels(b *testing.B) {
	const total = 6000
	ctx := context.Background()
	storage, hasher := newMemStorage(), newMemAccessHasher()
	if err := storage.SetState(ctx, 1, State{}); err != nil {
		b.Fatal(err)
	}
	for id := int64(1); id <= total; id++ {
		if err := storage.SetChannelPts(ctx, 1, id, 10); err != nil {
			b.Fatal(err)
		}
		if err := hasher.SetChannelAccessHash(ctx, 1, id, 123); err != nil {
			b.Fatal(err)
		}
	}
	for _, selected := range []bool{false, true} {
		name := "Default"
		if selected {
			name = "Selected20"
		}
		b.Run(name, func(b *testing.B) {
			cfg := Config{Storage: storage, AccessHasher: hasher,
				Handler: telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil }),
			}
			if selected {
				cfg.TrackChannel = func(id int64) bool { return id <= 20 }
			}
			m := New(cfg)
			b.ReportAllocs()
			for b.Loop() {
				channels, err := m.loadChannels(ctx, 1)
				if err != nil {
					b.Fatal(err)
				}
				b.ReportMetric(float64(len(channels)), "channels/op")
			}
		})
	}
}
