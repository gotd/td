package updates

import (
	"context"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

type gapState uint8

const (
	gapInit gapState = iota
	gapResolved
	gapRecover
)

type sequenceBox struct {
	state      int
	gaps       *gapBuffer
	gapTimeout *time.Timer
	pending    []update
	recovering bool

	apply     func(state int, updates []update) error
	notifyGap func(gapState)
	mux       sync.Mutex

	log *zap.Logger
}

type sequenceConfig struct {
	InitialState int
	Apply        func(state int, updates []update) error
	OnGap        func(gapState)
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
		log:       cfg.Logger,
	}
}

func (s *sequenceBox) Handle(u update) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.recovering {
		s.pending = append(s.pending, u)
		return nil
	}

	if s.gaps.Has() {
		s.pending = append(s.pending, u)
		if resolved := s.gaps.Consume(u); resolved {
			if s.gapTimeout.Stop() {
				s.notifyGap(gapResolved)
			}
			s.log.Debug("Gap was resolved by waiting")
			return s.applyPending()
		}

		s.log.Debug("Gap status", zap.Array("gaps", s.gaps))
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

		s.state = u.State
		return nil

	case gapIgnore:
		s.log.Debug("Old update", zap.Int("state", u.State))
		return nil

	case gapRefetch:
		s.pending = append(s.pending, u)
		s.gaps.Enable(s.state+1, u.start()-1)
		s.gapTimeout.Reset(fastgapTimeout)
		s.notifyGap(gapInit)
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
		cursor = i
		switch checkGap(state, update.State, update.Count) {
		case gapApply:
			accepted = append(accepted, update)
			state = update.State
			continue

		case gapIgnore:
			continue

		case gapRefetch:
			break loop
		}
	}

	if len(accepted) == 0 {
		s.log.Warn("Empty buffer", zap.Any("pending", s.pending), zap.Int("state", s.state))
		return nil
	}

	if err := s.apply(state, accepted); err != nil {
		return err
	}

	s.pending = s.pending[cursor+1:]
	s.state = state
	return nil
}

func (s *sequenceBox) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-s.gapTimeout.C:
			s.log.Info("Gap timeout")
			s.notifyGap(gapRecover)
		}
	}
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

func (s *sequenceBox) ExtractBuffer() []update {
	s.mux.Lock()
	defer s.mux.Unlock()
	defer func() { s.pending = nil }()
	return s.pending
}
