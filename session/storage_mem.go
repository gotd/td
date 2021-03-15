package session

import (
	"context"
	"sync"

	"golang.org/x/xerrors"
)

// StorageMemory implements in-memory session storage.
// Goroutine-safe.
type StorageMemory struct {
	mux  sync.RWMutex
	data []byte
}

// LoadSession loads session from memory.
func (s *StorageMemory) LoadSession(ctx context.Context) ([]byte, error) {
	if s == nil {
		return nil, ErrNotFound
	}
	s.mux.RLock()
	defer s.mux.RUnlock()

	if len(s.data) == 0 {
		return nil, ErrNotFound
	}
	cpy := append([]byte(nil), s.data...)

	return cpy, nil
}

// StoreSession stores session to memory.
func (s *StorageMemory) StoreSession(ctx context.Context, data []byte) error {
	if s == nil {
		return xerrors.New("StoreSession called on StorageMemory(nil)")
	}

	s.mux.Lock()
	s.data = data
	s.mux.Unlock()
	return nil
}
