// Package salts contains MTProto server salt storage.
package salts

import (
	"sort"
	"sync"
	"time"

	"github.com/gotd/td/internal/mt"
)

// Salts is a simple struct store server salts.
type Salts struct {
	// server salts fetched by getSalts.
	salts    []mt.FutureSalt
	saltsMux sync.Mutex
}

// Get returns next valid salt.
func (s *Salts) Get(buffer time.Duration) (int64, bool) {
	s.saltsMux.Lock()
	defer s.saltsMux.Unlock()

	// Sort slice by valid until.
	sort.SliceStable(s.salts, func(i, j int) bool {
		return s.salts[i].ValidUntil < s.salts[j].ValidUntil
	})

	// Filter (in place) from SliceTricks.
	n := 0
	dedup := map[int64]struct{}{}
	// Check that the salt will be valid next 5 minute.
	date := int(time.Now().Add(buffer).Unix())
	for _, salt := range s.salts {
		// Filter expired salts.
		if _, ok := dedup[salt.Salt]; !ok && salt.ValidUntil > date {
			dedup[salt.Salt] = struct{}{}
			s.salts[n] = salt
			n++
		}
	}
	s.salts = s.salts[:n]

	if len(s.salts) < 1 {
		return 0, false
	}
	return s.salts[0].Salt, true
}

// Store stores all given salts.
func (s *Salts) Store(salts []mt.FutureSalt) {
	s.saltsMux.Lock()
	s.salts = append(s.salts, salts...)
	s.saltsMux.Unlock()
}

// Reset deletes all stored salts.
func (s *Salts) Reset() {
	s.saltsMux.Lock()
	s.salts = s.salts[:0]
	s.saltsMux.Unlock()
}
