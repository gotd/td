package telegram

import (
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"net"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/tg"
)

// getVersion optimistically gets current client version.
//
// Does not handle replace directives.
func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	// Hard-coded package name. Probably we can generate this via parsing
	// the go.mod file.
	const pkg = "github.com/gotd/td"
	for _, d := range info.Deps {
		if strings.HasPrefix(d.Path, pkg) {
			return d.Version
		}
	}
	return ""
}

func parseCDNKeys(keys ...tg.CDNPublicKey) ([]*rsa.PublicKey, error) {
	r := make([]*rsa.PublicKey, 0, len(keys))

	for _, key := range keys {
		block, _ := pem.Decode([]byte(key.PublicKey))
		if block == nil {
			continue
		}

		key, err := crypto.ParseRSA(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA from PEM: %w", err)
		}

		r = append(r, key)
	}

	return r, nil
}

func (c *Client) lookupDC(id int) (tg.DCOption, error) {
	c.pmux.RLock()
	defer c.pmux.RUnlock()

	dc, ok := findDC(c.cfg, id, c.opts.PreferIPv6)
	if !ok {
		return tg.DCOption{}, xerrors.Errorf("dc not found in config: %d", id)
	}

	return dc, nil
}

func (c *Client) currentDC() (tg.DCOption, error) {
	addr, port, err := net.SplitHostPort(c.addr)
	if err != nil {
		return tg.DCOption{}, err
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		return tg.DCOption{}, err
	}

	return tg.DCOption{
		ID:        c.primaryDC,
		IPAddress: addr,
		Port:      p,
	}, nil
}

func findDC(cfg tg.Config, dcID int, preferIPv6 bool) (tg.DCOption, bool) {
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
		return tg.DCOption{}, false
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

	return opts[candidates[0]], true
}

func dcAttrs(dc tg.DCOption) (attrs []string) {
	if dc.CDN {
		attrs = append(attrs, "cdn")
	}
	if dc.MediaOnly {
		attrs = append(attrs, "media_only")
	}
	if dc.Static {
		attrs = append(attrs, "static")
	}
	if dc.TCPObfuscatedOnly {
		attrs = append(attrs, "tcpo")
	}
	return
}
