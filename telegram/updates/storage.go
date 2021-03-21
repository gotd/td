package updates

import (
	"context"
	"sync"
)

// Box is a abstraction for one storage entry.
type Box interface {
	Commit(ctx context.Context, pts int) error
	Load(ctx context.Context) (int, error)
}

// Storage is a abstraction for persistent storage.
type Storage interface {
	Acquire(ctx context.Context, name string, cb func(Box) error) error
	Keys(ctx context.Context) ([]string, error)
}

type box struct {
	Value int `json:"value"`
	mux   sync.Mutex
}

func (b *box) Commit(ctx context.Context, pts int) error {
	b.Value = pts
	return nil
}

func (b *box) Load(ctx context.Context) (int, error) {
	return b.Value, nil
}

// InMemory is a simple implementation of Storage.
type InMemory struct {
	kv  map[string]*box
	mux sync.Mutex
}

// Keys implements Storage.
func (s *InMemory) Keys(ctx context.Context) ([]string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	r := make([]string, 0, len(s.kv))
	for k := range s.kv {
		r = append(r, k)
	}

	return r, nil
}

// Acquire implements Storage.
func (s *InMemory) Acquire(ctx context.Context, name string, cb func(Box) error) error {
	s.mux.Lock()
	b, ok := s.kv[name]
	if !ok {
		b = &box{}
		s.kv[name] = b
	}
	s.mux.Unlock()

	b.mux.Lock()
	err := cb(b)
	b.mux.Unlock()
	return err
}

// Set can be used to set storage values manually.
func (s *InMemory) Set(name string, value int) {
	s.mux.Lock()
	b, ok := s.kv[name]
	if !ok {
		b = &box{}
		s.kv[name] = b
	}
	s.mux.Unlock()

	b.mux.Lock()
	b.Value = value
	b.mux.Unlock()
}

// NewInMemory creates new InMemory storage.
func NewInMemory() *InMemory {
	return &InMemory{kv: map[string]*box{}}
}
