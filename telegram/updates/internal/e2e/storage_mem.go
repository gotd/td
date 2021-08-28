package e2e

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/updates"
)

var _ updates.StateStorage = (*memStorage)(nil)

type memStorage struct {
	states   map[int]updates.State
	channels map[int]map[int]int
	mux      sync.Mutex
}

func newMemStorage() *memStorage {
	return &memStorage{
		states:   map[int]updates.State{},
		channels: map[int]map[int]int{},
	}
}

func (s *memStorage) GetState(userID int) (state updates.State, found bool, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, found = s.states[userID]
	return
}

func (s *memStorage) SetState(userID int, state updates.State) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.states[userID] = state
	s.channels[userID] = map[int]int{}
	return nil
}

func (s *memStorage) SetPts(userID, pts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return xerrors.Errorf("state not found")
	}

	state.Pts = pts
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetQts(userID, qts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return xerrors.Errorf("state not found")
	}

	state.Qts = qts
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetDate(userID, date int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return xerrors.Errorf("state not found")
	}

	state.Date = date
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetSeq(userID, seq int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return xerrors.Errorf("state not found")
	}

	state.Seq = seq
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetDateSeq(userID, date, seq int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	state, ok := s.states[userID]
	if !ok {
		return xerrors.Errorf("state not found")
	}

	state.Date = date
	state.Seq = seq
	s.states[userID] = state
	return nil
}

func (s *memStorage) SetChannelPts(userID, channelID, pts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	channels, ok := s.channels[userID]
	if !ok {
		return xerrors.Errorf("user state does not exist")
	}

	channels[channelID] = pts
	return nil
}

func (s *memStorage) GetChannelPts(userID, channelID int) (pts int, found bool, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	channels, ok := s.channels[userID]
	if !ok {
		return 0, false, nil
	}

	pts, found = channels[channelID]
	return
}

func (s *memStorage) ForEachChannels(userID int, f func(channelID, pts int) error) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	cmap, ok := s.channels[userID]
	if !ok {
		return xerrors.Errorf("channels map does not exist")
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
	hashes map[int]map[int]int64
	mux    sync.Mutex
}

func newMemAccessHasher() *memAccessHasher {
	return &memAccessHasher{
		hashes: map[int]map[int]int64{},
	}
}

func (m *memAccessHasher) GetChannelAccessHash(userID, channelID int) (hash int64, found bool, err error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	userHashes, ok := m.hashes[userID]
	if !ok {
		return 0, false, nil
	}

	hash, found = userHashes[channelID]
	return
}

func (m *memAccessHasher) SetChannelAccessHash(userID, channelID int, hash int64) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	userHashes, ok := m.hashes[userID]
	if !ok {
		userHashes = map[int]int64{}
		m.hashes[userID] = userHashes
	}

	userHashes[channelID] = hash
	return nil
}
