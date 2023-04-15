package updates

import (
	"context"
	"sync"

	"github.com/go-faster/errors"
)

var _ StateStorage = (*memStorage)(nil)

type memStorage struct {
	states   map[int64]State
	channels map[int64]map[int64]int
	mux      sync.Mutex
}

func newMemStorage() *memStorage {
	return &memStorage{
		states:   map[int64]State{},
		channels: map[int64]map[int64]int{},
	}
}

func (s *memStorage) GetState(ctx context.Context, userID int64) (state State, found bool, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, found = s.states[userID]
	return
}

func (s *memStorage) SetState(ctx context.Context, userID int64, state State) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.states[userID] = state
	s.channels[userID] = map[int64]int{}
	return nil
}

func (s *memStorage) SetPts(ctx context.Context, userID int64, pts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return errors.New("internalState not found")
	}

	state.Pts = pts
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetQts(ctx context.Context, userID int64, qts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return errors.New("internalState not found")
	}

	state.Qts = qts
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetDate(ctx context.Context, userID int64, date int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return errors.New("internalState not found")
	}

	state.Date = date
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetSeq(ctx context.Context, userID int64, seq int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return errors.New("internalState not found")
	}

	state.Seq = seq
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetDateSeq(ctx context.Context, userID int64, date, seq int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return errors.New("internalState not found")
	}

	state.Date = date
	state.Seq = seq
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetChannelPts(ctx context.Context, userID, channelID int64, pts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	channels, ok := s.channels[userID]
	if !ok {
		return errors.New("user internalState does not exist")
	}

	channels[channelID] = pts
	return nil
}

func (s *memStorage) GetChannelPts(ctx context.Context, userID, channelID int64) (pts int, found bool, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	channels, ok := s.channels[userID]
	if !ok {
		return 0, false, nil
	}

	pts, found = channels[channelID]
	return
}

func (s *memStorage) ForEachChannels(ctx context.Context, userID int64, f func(ctx context.Context, channelID int64, pts int) error) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	cmap, ok := s.channels[userID]
	if !ok {
		return errors.New("channels map does not exist")
	}

	for id, pts := range cmap {
		if err := f(ctx, id, pts); err != nil {
			return err
		}
	}

	return nil
}

var _ ChannelAccessHasher = (*memAccessHasher)(nil)

type memAccessHasher struct {
	hashes map[int64]map[int64]int64
	mux    sync.Mutex
}

func newMemAccessHasher() *memAccessHasher {
	return &memAccessHasher{
		hashes: map[int64]map[int64]int64{},
	}
}

func (m *memAccessHasher) GetChannelAccessHash(ctx context.Context, userID, channelID int64) (accessHash int64, found bool, err error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	userHashes, ok := m.hashes[userID]
	if !ok {
		return 0, false, nil
	}

	accessHash, found = userHashes[channelID]
	return
}

func (m *memAccessHasher) SetChannelAccessHash(ctx context.Context, userID, channelID, accessHash int64) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	userHashes, ok := m.hashes[userID]
	if !ok {
		userHashes = map[int64]int64{}
		m.hashes[userID] = userHashes
	}

	userHashes[channelID] = accessHash
	return nil
}
