package tdsync

import (
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type syncBackoff struct {
	b   backoff.BackOff
	mux sync.Mutex
}

func (s *syncBackoff) NextBackOff() time.Duration {
	s.mux.Lock()
	dur := s.b.NextBackOff()
	s.mux.Unlock()
	return dur
}

func (s *syncBackoff) Reset() {
	s.mux.Lock()
	s.b.Reset()
	s.mux.Unlock()
}

// SyncBackoff decorates backoff.BackOff to be thread-safe.
func SyncBackoff(from backoff.BackOff) backoff.BackOff {
	return &syncBackoff{b: from}
}
