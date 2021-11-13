package session

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/go-faster/errors"
)

// StorageMemory implements in-memory session storage.
// Goroutine-safe.
type StorageMemory struct {
	mux  sync.RWMutex
	data []byte
}

// Dump dumps raw session data to the given writer.
// Returns ErrNotFound if storage is nil or if underlying session is empty.
func (s *StorageMemory) Dump(w io.Writer) error {
	if s == nil {
		return ErrNotFound
	}
	s.mux.RLock()
	defer s.mux.RUnlock()

	if len(s.data) == 0 {
		return ErrNotFound
	}

	if _, err := w.Write(s.data); err != nil {
		return errors.Wrap(err, "write session")
	}

	return nil
}

// WriteFile dumps raw session data to the named file, creating it if necessary.
// Returns ErrNotFound if storage is nil or if underlying session is empty.
func (s *StorageMemory) WriteFile(name string, perm os.FileMode) error {
	data, err := s.Bytes(nil)
	if err != nil {
		return err
	}
	return os.WriteFile(name, data, perm)
}

// Bytes appends raw session data to the given slice.
// Returns ErrNotFound if storage is nil or if underlying session is empty.
func (s *StorageMemory) Bytes(to []byte) ([]byte, error) {
	if s == nil {
		return nil, ErrNotFound
	}
	s.mux.RLock()
	defer s.mux.RUnlock()

	if len(s.data) == 0 {
		return nil, ErrNotFound
	}

	return append(to, s.data...), nil
}

// Clone creates a clone of existing StorageMemory,
func (s *StorageMemory) Clone() *StorageMemory {
	s2 := &StorageMemory{}

	s2.data, _ = s.Bytes(s2.data)
	return s2
}

// LoadSession loads session from memory.
func (s *StorageMemory) LoadSession(context.Context) ([]byte, error) {
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
		return errors.New("StoreSession called on StorageMemory(nil)")
	}

	s.mux.Lock()
	s.data = data
	s.mux.Unlock()
	return nil
}
