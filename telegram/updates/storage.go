package updates

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// State contains common sequence state.
type State struct {
	Pts  int
	Qts  int
	Date int
	Seq  int
}

func (s State) fromRemote(remote *tg.UpdatesState) State {
	s.Pts = remote.Pts
	s.Qts = remote.Qts
	s.Date = remote.Date
	s.Seq = remote.Seq
	return s
}

// ErrStateNotFound means that state is not found in storage.
var ErrStateNotFound = xerrors.Errorf("state not found")

// Storage interface.
type Storage interface {
	GetState() (State, error)
	SetState(s State) error
	SetPts(pts int) error
	SetQts(qts int) error
	SetSeq(seq int) error
	SetDate(date int) error

	SetChannelPts(channelID, pts int) error
	Channels(iter func(channelID, pts int)) error

	ForgetAll() error
}

// MemStorage is a in-memory sequence storage.
type MemStorage struct {
	state     State
	haveState bool
	channels  map[int]int
	mux       sync.Mutex
}

// NewMemStorage creates new MemStorage.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		channels: map[int]int{},
	}
}

// GetState returns the state.
func (s *MemStorage) GetState() (State, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if !s.haveState {
		return State{}, ErrStateNotFound
	}
	return s.state, nil
}

// SetState sets the state.
func (s *MemStorage) SetState(state State) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.state = state
	s.haveState = true
	return nil
}

// SetPts sets pts.
func (s *MemStorage) SetPts(pts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.state.Pts = pts
	return nil
}

// SetQts sets qts.
func (s *MemStorage) SetQts(qts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.state.Qts = qts
	return nil
}

// SetSeq sets seq.
func (s *MemStorage) SetSeq(seq int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.state.Seq = seq
	return nil
}

// SetDate sets date.
func (s *MemStorage) SetDate(date int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.state.Date = date
	return nil
}

// SetChannelPts sets channel pts.
func (s *MemStorage) SetChannelPts(channelID, pts int) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.channels[channelID] = pts
	return nil
}

// Channels iterates through channels.
func (s *MemStorage) Channels(f func(channelID, pts int)) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	for channelID, pts := range s.channels {
		f(channelID, pts)
	}
	return nil
}

// ForgetAll clears all sequence info.
func (s *MemStorage) ForgetAll() error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.state = State{}
	s.channels = map[int]int{}
	return nil
}
