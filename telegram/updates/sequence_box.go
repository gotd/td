package updates

import (
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

type sequenceBox struct {
	state      int
	gaps       *gapBuffer
	gapTimeout *time.Timer
	pending    []update
	recovering bool

	apply     func(state int, updates []update) error
	notifyGap func()
	done      chan struct{}
	closed    bool
	mux       sync.Mutex
	wg        sync.WaitGroup

	log *zap.Logger
}

type sequenceConfig struct {
	InitialState int
	Apply        func(state int, updates []update) error
	OnGap        func()
	Logger       *zap.Logger
}

func newSequenceBox(cfg sequenceConfig) *sequenceBox {
	if cfg.Apply == nil {
		panic("Apply func nil")
	}
	if cfg.OnGap == nil {
		panic("OnGap func nil")
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}

	cfg.Logger.Debug("Initialized", zap.Int("state", cfg.InitialState))

	t := time.NewTimer(fastgapTimeout)
	t.Stop()
	return &sequenceBox{
		state:      cfg.InitialState,
		gaps:       new(gapBuffer),
		gapTimeout: t,

		apply:     cfg.Apply,
		notifyGap: cfg.OnGap,
		done:      make(chan struct{}),
		log:       cfg.Logger,
	}
}

func (s *sequenceBox) Handle(u update) error {
	s.mux.Lock()
	if s.closed {
		s.mux.Unlock()
		return xerrors.Errorf("closed")
	}
	defer s.mux.Unlock()

	log := s.log.With(zap.Int("upd_from", u.start()), zap.Int("upd_to", u.end()))
	if checkGap(s.state, u.State, u.Count) == gapIgnore {
		log.Debug("Outdated update, skip", zap.Int("state", s.state))
		return nil
	}

	if s.recovering {
		s.pending = append(s.pending, u)
		log.Debug("Postponed", zap.Int("pending_count", len(s.pending)))
		return nil
	}

	if s.gaps.Has() {
		s.pending = append(s.pending, u)
		accepted := s.gaps.Consume(u)
		if !accepted {
			log.Debug("Out of gap range, postponed", zap.Array("gaps", s.gaps))
			return nil
		}

		log.Debug("Gap accepted", zap.Array("gaps", s.gaps))
		if !s.gaps.Has() {
			_ = s.gapTimeout.Stop()
			s.log.Debug("Gap was resolved by waiting")
			return s.applyPending()
		}

		return nil
	}

	switch checkGap(s.state, u.State, u.Count) {
	case gapApply:
		if len(s.pending) > 0 {
			s.pending = append(s.pending, u)
			return s.applyPending()
		}

		if err := s.apply(u.State, []update{u}); err != nil {
			return err
		}

		log.Debug("Accepted")
		s.state = u.State
		return nil

	case gapRefetch:
		s.pending = append(s.pending, u)
		s.gaps.Enable(s.state+1, u.start()-1)

		// Check if we already have acceptable updates in buffer.
		for _, u := range s.pending {
			_ = s.gaps.Consume(u)
		}

		if !s.gaps.Has() {
			log.Debug("Gap was resolved by pending updates")
			return s.applyPending()
		}

		_ = s.gapTimeout.Reset(fastgapTimeout)
		s.log.Debug("Gap init", zap.Array("gap", s.gaps))
		return nil

	default:
		panic("unreachable")
	}
}

func (s *sequenceBox) applyPending() error {
	sort.SliceStable(s.pending, func(i, j int) bool {
		return s.pending[i].start() < s.pending[j].start()
	})

	var (
		cursor   = 0
		state    = s.state
		accepted []update
	)

loop:
	for i, update := range s.pending {
		switch checkGap(state, update.State, update.Count) {
		case gapApply:
			accepted = append(accepted, update)
			state = update.State
			cursor = i + 1
			continue

		case gapIgnore:
			cursor = i + 1
			continue

		case gapRefetch:
			break loop
		}
	}

	s.pending = s.pending[cursor:]
	if len(accepted) == 0 {
		s.log.Warn("Empty buffer", zap.Any("pending", s.pending), zap.Int("state", s.state))
		return nil
	}

	if err := s.apply(state, accepted); err != nil {
		return err
	}

	s.log.Debug("Pending updates applied",
		zap.Int("prev_state", s.state),
		zap.Int("new_state", state),
		zap.Int("updates_count", len(accepted)),
	)

	s.state = state
	return nil
}

func (s *sequenceBox) run() {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.done:
				return

			case <-s.gapTimeout.C:
				s.log.Info("Gap timeout")
				s.notifyGap()
			}
		}
	}()
}

func (s *sequenceBox) stop() {
	s.mux.Lock()
	defer s.mux.Unlock()
	close(s.done)
	s.closed = true
	s.wg.Wait()
}

func (s *sequenceBox) State() int {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.state
}

func (s *sequenceBox) SetState(state int) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.log.Debug("Forced set state", zap.Int("state", state))
	s.state = state
}

func (s *sequenceBox) EnableRecoverMode() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.recovering {
		panic("unreachable")
	}

	s.recovering = true
	s.gaps.Reset()
	s.gapTimeout.Stop()

	s.log.Debug("Recover mode enabled")
}

func (s *sequenceBox) DisableRecoverMode() {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.recovering {
		panic("unreachable")
	}

	s.recovering = false
	s.log.Debug("Recover mode disabled")
}
