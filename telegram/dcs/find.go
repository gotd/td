package dcs

import (
	"sort"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

func findDCs(cfg tg.Config, dcID int, preferIPv6 bool) ([]int, bool) {
	// Preallocate slice.
	candidates := make([]int, 0, 32)

	opts := cfg.DCOptions
	for idx, candidateDC := range opts {
		if candidateDC.ID != dcID {
			continue
		}
		candidates = append(candidates, idx)
	}

	if len(candidates) < 1 {
		return nil, false
	}

	sort.Slice(candidates, func(i, j int) bool {
		l, r := opts[candidates[i]], opts[candidates[j]]

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

	return candidates, true
}

// FindDC searches DC from given config.
func FindDC(cfg tg.Config, dcID int, preferIPv6 bool) (tg.DCOption, bool) {
	candidates, ok := findDCs(cfg, dcID, preferIPv6)
	if !ok {
		return tg.DCOption{}, false
	}

	return cfg.DCOptions[candidates[0]], true
}

// FindPrimaryDC searches new primary DC from given config.
// Unlike FindDC, it filters CDNs and MediaOnly servers, returns error
// if not found.
func FindPrimaryDC(cfg tg.Config, dcID int, preferIPv6 bool) (tg.DCOption, error) {
	candidates, ok := findDCs(cfg, dcID, preferIPv6)
	if !ok {
		return tg.DCOption{}, xerrors.Errorf("can't find DC %d", dcID)
	}
	opts := cfg.DCOptions

	// Filter (in place) from SliceTricks.
	n := 0
	for _, idx := range candidates {
		opt := opts[idx]
		if !opt.MediaOnly && !opt.CDN && !opt.TCPObfuscatedOnly {
			candidates[n] = idx
			n++
		}
	}
	candidates = candidates[:n]

	if len(candidates) < 1 {
		return tg.DCOption{}, xerrors.Errorf("find new primary DC %d", dcID)
	}
	return opts[candidates[0]], nil
}
