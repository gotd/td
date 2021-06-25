package updates

import (
	"sync"

	"go.uber.org/zap"
)

// AccessHasher stores channel access hashes.
type AccessHasher interface {
	GetChannelHash(channelID int) (hash int64, found bool, err error)
	SetChannelHash(channelID int, accessHash int64) error
}

type memAccessHasher struct {
	hashes map[int]int64
	mux    sync.Mutex
}

func newMemAccessHasher() *memAccessHasher {
	return &memAccessHasher{hashes: map[int]int64{}}
}

func (m *memAccessHasher) GetChannelHash(channelID int) (hash int64, found bool, err error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	hash, found = m.hashes[channelID]
	return
}

func (m *memAccessHasher) SetChannelHash(channelID int, accessHash int64) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.hashes[channelID] = accessHash
	return nil
}

type hashStorage struct {
	hasher AccessHasher
	log    *zap.Logger
}

func (s *hashStorage) Get(channelID int) (int64, bool) {
	hash, found, err := s.hasher.GetChannelHash(channelID)
	if err != nil {
		s.log.Error("Failed to get access hash", zap.Int("channel_id", channelID), zap.Error(err))
		return 0, false
	}

	return hash, found
}

func (s *hashStorage) Set(channelID int, hash int64) {
	if err := s.hasher.SetChannelHash(channelID, hash); err != nil {
		s.log.Error("Failed to save access hash", zap.Int("channel_id", channelID), zap.Error(err))
	}
}
