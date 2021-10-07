package e2e

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/updates"
)

var _ updates.StateStorage = (*memStorage)(nil)

type memStorage struct {
	states   map[int64]updates.State
	channels map[int64]map[int64]int
	mux      sync.Mutex
}

func newMemStorage() *memStorage {
	return &memStorage{
		states:   map[int64]updates.State{},
		channels: map[int64]map[int64]int{},
	}
}

func (s *memStorage) GetState(userID int64) (state updates.State, found bool, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, found = s.states[userID]
	return
}

func (s *memStorage) SetState(userID int64, state updates.State) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.states[userID] = state
	s.channels[userID] = map[int64]int{}
	return nil
}

func (s *memStorage) SetPts(userID int64, pts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return xerrors.New("state not found")
	}

	state.Pts = pts
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetQts(userID int64, qts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return xerrors.New("state not found")
	}

	state.Qts = qts
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetDate(userID int64, date int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return xerrors.New("state not found")
	}

	state.Date = date
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetSeq(userID int64, seq int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return xerrors.New("state not found")
	}

	state.Seq = seq
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetDateSeq(userID int64, date, seq int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return xerrors.New("state not found")
	}

	state.Date = date
	state.Seq = seq
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetChannelPts(userID, channelID int64, pts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	channels, ok := s.channels[userID]
	if !ok {
		return xerrors.New("user state does not exist")
	}

	channels[channelID] = pts
	return nil
}

func (s *memStorage) GetChannelPts(userID, channelID int64) (pts int, found bool, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	channels, ok := s.channels[userID]
	if !ok {
		return 0, false, nil
	}

	pts, found = channels[channelID]
	return
}

func (s *memStorage) ForEachChannels(userID int64, f func(channelID int64, pts int) error) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	cmap, ok := s.channels[userID]
	if !ok {
		return xerrors.New("channels map does not exist")
	}

	for id, pts := range cmap {
		if err := f(id, pts); err != nil {
			return err
		}
	}

	return nil
}

var _ updates.ChannelAccessHasher = (*memAccessHasher)(nil)

type memAccessHasher struct {
	hashes map[int64]map[int64]int64
	mux    sync.Mutex
}

func newMemAccessHasher() *memAccessHasher {
	return &memAccessHasher{
		hashes: map[int64]map[int64]int64{},
	}
}

func (m *memAccessHasher) GetChannelAccessHash(userID, channelID int64) (hash int64, found bool, err error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	userHashes, ok := m.hashes[userID]
	if !ok {
		return 0, false, nil
	}

	hash, found = userHashes[channelID]
	return
}

func (m *memAccessHasher) SetChannelAccessHash(userID, channelID, hash int64) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	userHashes, ok := m.hashes[userID]
	if !ok {
		userHashes = map[int64]int64{}
		m.hashes[userID] = userHashes
	}

	userHashes[channelID] = hash
	return nil
}
