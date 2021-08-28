package updates

import (
	"sort"
	"time"

	"go.uber.org/zap"
)

type sequenceBox struct {
	state      int
	gaps       *gapBuffer
	gapTimeout *time.Timer
	pending    []update

	apply func(state int, updates []update) error
	log   *zap.Logger
}

type sequenceConfig struct {
	InitialState int
	Apply        func(state int, updates []update) error
	Logger       *zap.Logger
}

func newSequenceBox(cfg sequenceConfig) *sequenceBox {
	if cfg.Apply == nil {
		panic("Apply func nil")
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}

	cfg.Logger.Debug("Initialized", zap.Int("state", cfg.InitialState))

	t := time.NewTimer(fastgapTimeout)
	_ = t.Stop()
	return &sequenceBox{
		state:      cfg.InitialState,
		gaps:       new(gapBuffer),
		gapTimeout: t,

		apply: cfg.Apply,
		log:   cfg.Logger,
	}
}

func (s *sequenceBox) Handle(u update) error {
	log := s.log.With(zap.Int("upd_from", u.start()), zap.Int("upd_to", u.end()))
	if checkGap(s.state, u.State, u.Count) == gapIgnore {
		log.Debug("Outdated update, skip", zap.Int("state", s.state))
		return nil
	}

	if s.gaps.Has() {
		s.pending = append(s.pending, u)
		if accepted := s.gaps.Consume(u); !accepted {
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
		s.gaps.Enable(s.state, u.start())

		// Check if we already have acceptable updates in buffer.
		for _, u := range s.pending {
			_ = s.gaps.Consume(u)
		}

		if !s.gaps.Has() {
			log.Debug("Gap was resolved by pending updates")
			return s.applyPending()
		}

		_ = s.gapTimeout.Reset(fastgapTimeout)
		s.log.Debug("Gap detected", zap.Array("gap", s.gaps))
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

	// Trim processed updates. Setting zero values for the rest
	// of the slice lets GC collect referenced objects.
	end := len(s.pending)
	trim := end - cursor
	copy(s.pending, s.pending[cursor:])
	for i := trim; i < end; i++ {
		s.pending[i] = update{}
	}
	s.pending = s.pending[:trim]
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
		zap.Int("accepted_count", len(accepted)),
	)

	s.state = state
	return nil
}

func (s *sequenceBox) State() int { return s.state }

func (s *sequenceBox) SetState(state int) {
	s.log.Debug("Forced set state", zap.Int("state", state))
	s.state = state
}
