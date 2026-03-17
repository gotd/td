package telegram

import (
	"context"
	"crypto/rsa"
	"encoding/pem"

	"github.com/go-faster/errors"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/tg"
)

// Single key because help.getCdnConfig has no request params.
const helpGetCDNConfigSingleflightKey = "help.getCdnConfig"

type cdnKeyEntry struct {
	dcID int
	key  *rsa.PublicKey
}

type fetchedCDNKeys struct {
	all  []exchange.PublicKey
	byDC map[int][]exchange.PublicKey
}

func clonePublicKeys(keys []exchange.PublicKey) []exchange.PublicKey {
	return append([]exchange.PublicKey(nil), keys...)
}

func mergePublicKeys(primary, fallback []exchange.PublicKey) []exchange.PublicKey {
	if len(primary) == 0 && len(fallback) == 0 {
		return nil
	}

	out := make([]exchange.PublicKey, 0, len(primary)+len(fallback))
	seen := make(map[int64]struct{}, len(primary)+len(fallback))
	appendUnique := func(keys []exchange.PublicKey) {
		for _, key := range keys {
			fp := key.Fingerprint()
			if _, ok := seen[fp]; ok {
				continue
			}
			seen[fp] = struct{}{}
			out = append(out, key)
		}
	}

	// Prefer primary keyset order and use fallback only for missing fingerprints.
	appendUnique(primary)
	appendUnique(fallback)
	return out
}

func parseCDNKeyEntries(keys ...tg.CDNPublicKey) ([]cdnKeyEntry, error) {
	r := make([]cdnKeyEntry, 0, len(keys))

	for _, key := range keys {
		block, _ := pem.Decode([]byte(key.PublicKey))
		if block == nil {
			continue
		}

		parsedKey, err := crypto.ParseRSA(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "parse RSA from PEM")
		}

		r = append(r, cdnKeyEntry{
			dcID: key.DCID,
			key:  parsedKey,
		})
	}

	return r, nil
}

func buildCDNKeysCache(entries []cdnKeyEntry) fetchedCDNKeys {
	result := fetchedCDNKeys{
		all:  make([]exchange.PublicKey, 0, len(entries)),
		byDC: make(map[int][]exchange.PublicKey),
	}

	seenAll := make(map[int64]struct{}, len(entries))
	seenByDC := make(map[int]map[int64]struct{})

	for _, entry := range entries {
		key := exchange.PublicKey{RSA: entry.key}
		fingerprint := key.Fingerprint()

		if _, ok := seenAll[fingerprint]; !ok {
			seenAll[fingerprint] = struct{}{}
			result.all = append(result.all, key)
		}

		seen, ok := seenByDC[entry.dcID]
		if !ok {
			seen = map[int64]struct{}{}
			seenByDC[entry.dcID] = seen
		}
		if _, ok := seen[fingerprint]; ok {
			continue
		}
		seen[fingerprint] = struct{}{}
		result.byDC[entry.dcID] = append(result.byDC[entry.dcID], key)
	}

	return result
}

func copyCDNKeysByDC(byDC map[int][]exchange.PublicKey) map[int][]exchange.PublicKey {
	if len(byDC) == 0 {
		return nil
	}

	r := make(map[int][]exchange.PublicKey, len(byDC))
	for dcID, keys := range byDC {
		r[dcID] = append([]exchange.PublicKey(nil), keys...)
	}
	return r
}

func cloneFetchedCDNKeys(keys fetchedCDNKeys) fetchedCDNKeys {
	return fetchedCDNKeys{
		all:  clonePublicKeys(keys.all),
		byDC: copyCDNKeysByDC(keys.byDC),
	}
}

func (c *Client) cachedCDNKeys() ([]exchange.PublicKey, bool, uint64) {
	c.cdnKeysMux.Lock()
	defer c.cdnKeysMux.Unlock()

	return clonePublicKeys(c.cdnKeys), c.cdnKeysSet, c.cdnKeysGen
}

func (c *Client) cachedCDNKeysForDC(dcID int) ([]exchange.PublicKey, bool) {
	c.cdnKeysMux.Lock()
	defer c.cdnKeysMux.Unlock()

	return clonePublicKeys(c.cdnKeysByDC[dcID]), c.cdnKeysSet
}

func (c *Client) cdnConfigFetchContext(caller context.Context) context.Context {
	if c.ctx != nil {
		// Bind network request lifetime to client lifecycle, not to the first
		// singleflight caller.
		return c.ctx
	}

	// Caller cancellation is handled outside singleflight wait loop; request
	// itself should not inherit first caller deadline/cancellation.
	return context.WithoutCancel(caller)
}

func (c *Client) loadCDNKeys(ctx context.Context) (fetchedCDNKeys, error) {
	resultCh := c.cdnKeysLoad.DoChan(helpGetCDNConfigSingleflightKey, func() (interface{}, error) {
		// singleflight ensures only one goroutine issues help.getCdnConfig;
		// others wait and reuse same result.
		cfg, err := c.tg.HelpGetCDNConfig(c.cdnConfigFetchContext(ctx))
		if err != nil {
			return nil, errors.Wrap(err, "help.getCdnConfig")
		}

		entries, err := parseCDNKeyEntries(cfg.PublicKeys...)
		if err != nil {
			return nil, errors.Wrap(err, "parse CDN public keys")
		}
		return buildCDNKeysCache(entries), nil
	})

	select {
	case <-ctx.Done():
		return fetchedCDNKeys{}, ctx.Err()
	case result := <-resultCh:
		if result.Err != nil {
			return fetchedCDNKeys{}, result.Err
		}

		keys, ok := result.Val.(fetchedCDNKeys)
		if !ok {
			return fetchedCDNKeys{}, errors.Errorf("unexpected CDN keys type %T", result.Val)
		}
		return cloneFetchedCDNKeys(keys), nil
	}
}

func (c *Client) fetchCDNKeys(ctx context.Context) ([]exchange.PublicKey, error) {
	const maxVersionRetries = 3
	for attempt := 0; attempt < maxVersionRetries; attempt++ {
		// Fast path: fully cached, no network requests.
		cached, set, startGen := c.cachedCDNKeys()
		if set {
			return cached, nil
		}
		// Snapshot generation to detect invalidation races after in-flight load.

		keys, err := c.loadCDNKeys(ctx)
		if err != nil {
			return nil, err
		}

		c.cdnKeysMux.Lock()
		switch {
		case c.cdnKeysSet:
			// Another goroutine already populated cache while we were waiting.
			cached := clonePublicKeys(c.cdnKeys)
			c.cdnKeysMux.Unlock()
			return cached, nil
		case c.cdnKeysGen != startGen:
			// Cache was invalidated (fingerprint miss) during in-flight request.
			// Discard stale result and retry from fresh generation.
			c.cdnKeysMux.Unlock()
			continue
		default:
			// Safe to commit fetched keys into cache.
			c.cdnKeys = clonePublicKeys(keys.all)
			c.cdnKeysByDC = copyCDNKeysByDC(keys.byDC)
			c.cdnKeysSet = true
			cached := clonePublicKeys(c.cdnKeys)
			c.cdnKeysMux.Unlock()
			return cached, nil
		}
	}

	return nil, errors.New("cdn keys cache changed concurrently")
}

func (c *Client) refreshCDNKeys(ctx context.Context) ([]exchange.PublicKey, error) {
	const maxVersionRetries = 3
	for attempt := 0; attempt < maxVersionRetries; attempt++ {
		c.cdnKeysMux.Lock()
		startGen := c.cdnKeysGen
		c.cdnKeysMux.Unlock()

		keys, err := c.loadCDNKeys(ctx)
		if err != nil {
			return nil, err
		}

		c.cdnKeysMux.Lock()
		if c.cdnKeysGen != startGen {
			// Fingerprint invalidation happened while refresh was in-flight.
			// Discard stale result and refetch for fresh generation.
			c.cdnKeysMux.Unlock()
			continue
		}
		c.cdnKeys = clonePublicKeys(keys.all)
		c.cdnKeysByDC = copyCDNKeysByDC(keys.byDC)
		c.cdnKeysSet = true
		cached := clonePublicKeys(c.cdnKeys)
		c.cdnKeysMux.Unlock()

		return cached, nil
	}

	return nil, errors.New("cdn keys cache changed concurrently")
}

func (c *Client) fetchCDNKeysForDC(ctx context.Context, dcID int) ([]exchange.PublicKey, error) {
	keys, set := c.cachedCDNKeysForDC(dcID)
	if !set {
		if _, err := c.fetchCDNKeys(ctx); err != nil {
			return nil, err
		}
	}

	const maxRefreshAttempts = 3
	for attempt := 0; attempt < maxRefreshAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		keys, _ = c.cachedCDNKeysForDC(dcID)
		if len(keys) > 0 {
			return keys, nil
		}
		if attempt == maxRefreshAttempts-1 {
			break
		}

		// Requested CDN DC is missing in current snapshot; retry bounded
		// help.getCdnConfig refreshes to handle eventual config propagation.
		if _, err := c.refreshCDNKeys(ctx); err != nil {
			return nil, err
		}
	}

	return nil, errors.Errorf("no CDN public keys for CDN DC %d after %d refresh attempts", dcID, maxRefreshAttempts)
}
