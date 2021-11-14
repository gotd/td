package dcs

import (
	"sort"

	"github.com/gotd/td/tg"
)

// FindDCs searches DCs candidates from given config.
func FindDCs(opts []tg.DCOption, dcID int, preferIPv6 bool) []tg.DCOption {
	// Preallocate slice.
	candidates := make([]tg.DCOption, 0, 32)

	for _, candidateDC := range opts {
		if candidateDC.ID != dcID {
			continue
		}
		candidates = append(candidates, candidateDC)
	}

	if len(candidates) < 1 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		l, r := candidates[i], candidates[j]

		// If we prefer IPv6 and left is IPv6 and right is not, so then
		// left is smaller (would be before right).
		if preferIPv6 {
			if l.Ipv6 && !r.Ipv6 {
				return true
			}
			if !l.Ipv6 && r.Ipv6 {
				return false
			}
		}

		// Also we prefer static addresses.
		return l.Static && !r.Static
	})

	return candidates
}

// FindPrimaryDCs searches new primary DC from given config.
// Unlike FindDC, it filters CDNs and MediaOnly servers, returns error
// if not found.
func FindPrimaryDCs(opts []tg.DCOption, dcID int, preferIPv6 bool) []tg.DCOption {
	candidates := FindDCs(opts, dcID, preferIPv6)
	// Filter (in place) from SliceTricks.
	n := 0
	for _, opt := range candidates {
		if !opt.MediaOnly && !opt.CDN {
			candidates[n] = opt
			n++
		}
	}
	return candidates[:n]
}
