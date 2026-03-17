package downloader

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"sort"
	"strconv"

	"github.com/go-faster/errors"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/tg"
)

func (c *cdn) resetHashes() {
	c.hashesMux.Lock()
	c.hashes = nil
	c.hashOffsets = nil
	c.hashesMux.Unlock()
}

func (c *cdn) resetWindows() {
	c.windowsMux.Lock()
	// Drop all previously verified window payloads whenever redirect/token scope
	// changes to avoid mixing data that belongs to different CDN contexts.
	c.windows = nil
	c.windowsFIFO = nil
	c.windowsMux.Unlock()
}

func (c *cdn) cachedWindow(hash tg.FileHash) ([]byte, bool) {
	c.windowsMux.Lock()
	defer c.windowsMux.Unlock()

	data, ok := c.windows[hash.Offset]
	if !ok {
		return nil, false
	}
	if len(data) == 0 || len(data) > hash.Limit {
		return nil, false
	}

	// Returned slice is treated as read-only by callers.
	return data, true
}

func (c *cdn) cacheWindow(hash tg.FileHash, data []byte) {
	if hash.Limit <= 0 || len(data) == 0 || len(data) > hash.Limit {
		return
	}

	c.windowsMux.Lock()
	defer c.windowsMux.Unlock()
	if c.windows == nil {
		c.windows = make(map[int64][]byte)
	}
	if _, ok := c.windows[hash.Offset]; !ok {
		c.windowsFIFO = append(c.windowsFIFO, hash.Offset)
	}
	// Store a copy to keep cache immutable relative to caller buffers.
	c.windows[hash.Offset] = append([]byte(nil), data...)
	for len(c.windowsFIFO) > maxVerifiedWindowCache {
		evict := c.windowsFIFO[0]
		c.windowsFIFO = c.windowsFIFO[1:]
		// FIFO eviction is enough here: access pattern is near-sequential and we
		// only need to cap memory, not optimize for perfect hit rate.
		delete(c.windows, evict)
	}
}

func (c *cdn) cacheHashes(hashes []tg.FileHash) {
	if len(hashes) == 0 {
		return
	}

	c.hashesMux.Lock()
	if c.hashes == nil {
		c.hashes = make(map[int64]tg.FileHash, len(hashes))
	}
	for _, hash := range hashes {
		if hash.Limit <= 0 {
			continue
		}

		// Keep sorted unique offsets index for O(log n) range lookup.
		if _, exists := c.hashes[hash.Offset]; !exists {
			idx := sort.Search(len(c.hashOffsets), func(i int) bool {
				return c.hashOffsets[i] >= hash.Offset
			})
			if idx == len(c.hashOffsets) {
				c.hashOffsets = append(c.hashOffsets, hash.Offset)
			} else if c.hashOffsets[idx] != hash.Offset {
				c.hashOffsets = append(c.hashOffsets, 0)
				copy(c.hashOffsets[idx+1:], c.hashOffsets[idx:])
				c.hashOffsets[idx] = hash.Offset
			}
		}
		c.hashes[hash.Offset] = hash
	}
	c.hashesMux.Unlock()
}

func (c *cdn) hash(offset int64) (tg.FileHash, bool) {
	c.hashesMux.RLock()
	hash, ok := c.hashes[offset]
	if ok {
		c.hashesMux.RUnlock()
		return hash, true
	}

	// Fast-path map lookup works only for exact hash offsets. For unaligned
	// part sizes resolve containing window by predecessor offset in sorted index.
	if len(c.hashOffsets) == 0 {
		c.hashesMux.RUnlock()
		return tg.FileHash{}, false
	}

	idx := sort.Search(len(c.hashOffsets), func(i int) bool {
		return c.hashOffsets[i] > offset
	}) - 1
	if idx < 0 {
		c.hashesMux.RUnlock()
		return tg.FileHash{}, false
	}

	candidate, exists := c.hashes[c.hashOffsets[idx]]
	if !exists || candidate.Limit <= 0 {
		c.hashesMux.RUnlock()
		return tg.FileHash{}, false
	}
	end := candidate.Offset + int64(candidate.Limit)
	hash = candidate
	ok = offset >= candidate.Offset && offset < end
	c.hashesMux.RUnlock()
	return hash, ok
}

func (c *cdn) hashForOffset(ctx context.Context, offset int64) (tg.FileHash, error) {
	if hash, ok := c.hash(offset); ok {
		return hash, nil
	}

	// Ask server for the current offset window and cache returned range.
	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return tg.FileHash{}, err
		}

		hashes, err := c.Hashes(ctx, offset)
		if err != nil {
			return tg.FileHash{}, errors.Wrapf(err, "load CDN hashes at offset=%d", offset)
		}
		// Cache batch and retry lookup: server may return a range of windows
		// where requested offset is not the first element.
		c.cacheHashes(hashes)
		if hash, ok := c.hash(offset); ok {
			return hash, nil
		}
	}

	return tg.FileHash{}, retryLimitErr(
		"cdn hash lookup",
		maxRetryAttempts,
		errors.Errorf("hash for offset %d not found", offset),
	)
}

func windowLoadKey(hash tg.FileHash) string {
	key := make([]byte, 0, len(hash.Hash)+64)
	key = strconv.AppendInt(key, hash.Offset, 10)
	key = append(key, ':')
	key = strconv.AppendInt(key, int64(hash.Limit), 10)
	key = append(key, ':')
	key = append(key, hash.Hash...)
	return string(key)
}

func (c *cdn) loadAndVerifyWindow(ctx context.Context, hash tg.FileHash) ([]byte, error) {
	if window, ok := c.cachedWindow(hash); ok {
		return window, nil
	}

	key := windowLoadKey(hash)
	v, err, _ := c.windowsLoad.Do(key, func() (interface{}, error) {
		if window, ok := c.cachedWindow(hash); ok {
			return window, nil
		}

		// Fetching a whole window can return a shorter payload only for the
		// last file segment, so we accept len(full.data) <= hash.Limit.
		full, err := c.Chunk(ctx, hash.Offset, hash.Limit)
		if err != nil {
			return nil, errors.Wrapf(err, "load full hash window at offset=%d limit=%d", hash.Offset, hash.Limit)
		}
		if len(full.data) == 0 || len(full.data) > hash.Limit {
			return nil, errors.Errorf(
				"invalid CDN window length at offset=%d max=%d got=%d",
				hash.Offset, hash.Limit, len(full.data),
			)
		}
		if !bytes.Equal(crypto.SHA256(full.data), hash.Hash) {
			return nil, errors.Wrapf(
				ErrHashMismatch,
				"at offset=%d size=%d",
				hash.Offset, hash.Limit,
			)
		}

		c.cacheWindow(hash, full.data)
		return full.data, nil
	})
	if err != nil {
		return nil, err
	}

	window, ok := v.([]byte)
	if !ok {
		return nil, errors.Errorf("unexpected window type %T", v)
	}
	return window, nil
}

func (c *cdn) verifyChunk(ctx context.Context, offset int64, requestedLimit int, data []byte) error {
	if !c.verify || len(data) == 0 {
		return nil
	}
	shortResponse := requestedLimit > 0 && len(data) < requestedLimit

	// Inline mode validates every hash window covered by this chunk.
	// For windows split by custom part sizes, we load and verify the full window
	// (cached) and then patch overlapping bytes in this chunk with verified data.
	chunkStart := offset
	chunkEnd := offset + int64(len(data))
	for current := chunkStart; current < chunkEnd; {
		hash, err := c.hashForOffset(ctx, current)
		if err != nil {
			return err
		}
		if hash.Limit <= 0 {
			return errors.Errorf("invalid CDN hash limit %d at offset %d", hash.Limit, current)
		}
		windowStart := hash.Offset
		windowEnd := hash.Offset + int64(hash.Limit)
		if windowEnd <= current {
			return errors.Errorf("invalid CDN hash window [%d,%d) at offset %d", windowStart, windowEnd, current)
		}

		switch {
		case windowStart >= chunkStart && windowEnd <= chunkEnd:
			// Full hash window is present in current chunk: verify directly.
			from := int(windowStart - chunkStart)
			to := int(windowEnd - chunkStart)
			if !bytes.Equal(crypto.SHA256(data[from:to]), hash.Hash) {
				return errors.Wrapf(
					ErrHashMismatch,
					"at offset=%d size=%d",
					windowStart, hash.Limit,
				)
			}

		case shortResponse && windowStart >= chunkStart && windowStart < chunkEnd && windowEnd > chunkEnd:
			// Final short chunk: Telegram keeps nominal hash limit, but hash is
			// computed on actual remaining tail bytes.
			from := int(windowStart - chunkStart)
			if !bytes.Equal(crypto.SHA256(data[from:]), hash.Hash) {
				return ErrHashMismatch
			}
			return nil

		default:
			// Hash window crosses current chunk boundaries.
			//
			// TDesktop-style behavior: validate complete hash window and then apply
			// verified overlap to current chunk. This preserves integrity checks for
			// custom part sizes without forcing eager verifier mode globally.
			window, err := c.loadAndVerifyWindow(ctx, hash)
			if err != nil {
				return err
			}

			overlapStart := chunkStart
			if windowStart > overlapStart {
				overlapStart = windowStart
			}
			windowDataEnd := windowStart + int64(len(window))
			overlapEnd := chunkEnd
			if windowDataEnd < overlapEnd {
				overlapEnd = windowDataEnd
			}
			if overlapEnd <= overlapStart {
				return errors.Errorf(
					"invalid overlap for hash window [%d,%d) and chunk [%d,%d)",
					windowStart, windowEnd, chunkStart, chunkEnd,
				)
			}
			chunkFrom := int(overlapStart - chunkStart)
			chunkTo := int(overlapEnd - chunkStart)
			windowFrom := int(overlapStart - windowStart)
			windowTo := int(overlapEnd - windowStart)
			// Replace bytes in-place with verified overlap so the caller receives a
			// fully verified chunk even when hash windows are split by part size.
			copy(data[chunkFrom:chunkTo], window[windowFrom:windowTo])
		}
		current = windowEnd
	}
	return nil
}

// decrypt decrypts file chunk from Telegram CDN.
// See https://core.telegram.org/cdn#decrypting-files.
func (c *cdn) decrypt(src []byte, offset int64, redirect *tg.UploadFileCDNRedirect) ([]byte, error) {
	block, err := aes.NewCipher(redirect.EncryptionKey)
	if err != nil {
		return nil, errors.Wrap(err, "create cipher")
	}
	if block.BlockSize() != len(redirect.EncryptionIv) {
		return nil, errors.Errorf(
			"invalid IV or key length, block size %d != IV %d",
			block.BlockSize(), len(redirect.EncryptionIv),
		)
	}

	iv := c.pool.GetSize(len(redirect.EncryptionIv))
	defer c.pool.Put(iv)
	copy(iv.Buf, redirect.EncryptionIv)

	binary.BigEndian.PutUint32(iv.Buf[iv.Len()-4:], uint32(offset/16))

	dst := make([]byte, len(src))
	cipher.NewCTR(block, iv.Buf).XORKeyStream(dst, src)
	return dst, nil
}
