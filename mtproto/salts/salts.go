// Package salts contains MTProto server salt storage.
package salts

import (
	"sort"
	"sync"
	"time"

	"github.com/gotd/td/mt"
)

// Salts is a simple struct store server salts.
type Salts struct {
	// server salts fetched by getSalts.
	salts    []mt.FutureSalt
	saltsMux sync.Mutex
}

// Get returns next valid salt.
func (s *Salts) Get(deadline time.Time) (int64, bool) {
	s.saltsMux.Lock()
	defer s.saltsMux.Unlock()

check:
	if len(s.salts) < 1 {
		return 0, false
	}

	date := int(deadline.Unix())
	if salt := s.salts[len(s.salts)-1]; salt.ValidUntil > date {
		return salt.Salt, true
	}

	// Filter (in place) from SliceTricks.
	n := 0
	// Check that the salt will be valid until deadline.
	for _, salt := range s.salts {
		// Filter expired salts.
		if salt.ValidUntil > date {
			// Keep valid salt.
			s.salts[n] = salt
			n++
		}
	}
	s.salts = s.salts[:n]
	goto check
}

type saltSlice []mt.FutureSalt

func (s saltSlice) Len() int {
	return len(s)
}

func (s saltSlice) Less(i, j int) bool {
	return s[i].ValidUntil > s[j].ValidUntil
}

func (s saltSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Store stores all given salts.
func (s *Salts) Store(salts []mt.FutureSalt) {
	s.saltsMux.Lock()
	defer s.saltsMux.Unlock()

	s.salts = append(s.salts, salts...)
	// Filter duplicates.
	n := 0
	dedup := make(map[int64]struct{}, len(s.salts)+1)
	for _, salt := range s.salts {
		if _, ok := dedup[salt.Salt]; !ok {
			dedup[salt.Salt] = struct{}{}
			s.salts[n] = salt
			n++
		}
	}
	s.salts = s.salts[:n]

	// Sort slice by valid until.
	sort.Sort(saltSlice(s.salts))
}

// Reset deletes all stored salts.
func (s *Salts) Reset() {
	s.saltsMux.Lock()
	s.salts = s.salts[:0]
	s.saltsMux.Unlock()
}
