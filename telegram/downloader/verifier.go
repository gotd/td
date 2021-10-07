package downloader

import (
	"bytes"
	"context"
	"sort"
	"sync"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
)

// ErrHashMismatch means that download hash verification was failed.
var ErrHashMismatch = xerrors.New("file hash mismatch")

type verifier struct {
	client schema

	hashes []tg.FileHash
	offset int
	mux    sync.Mutex
}

func newVerifier(client schema, hashes ...tg.FileHash) *verifier {
	r := make([]tg.FileHash, len(hashes))

	copy(r, hashes)
	sort.SliceStable(r, func(i, j int) bool {
		return r[i].Offset < r[j].Offset
	})

	return &verifier{client: client, hashes: r}
}

func (v *verifier) pop() (tg.FileHash, bool) {
	if len(v.hashes) < 1 {
		return tg.FileHash{}, false
	}

	// Pop and move.
	hash := v.hashes[0]
	copy(v.hashes, v.hashes[1:])
	v.hashes[len(v.hashes)-1] = tg.FileHash{}
	v.hashes = v.hashes[:len(v.hashes)-1]

	return hash, true
}

func (v *verifier) update(hashes ...tg.FileHash) (tg.FileHash, bool) {
	// If result is empty and queue is empty, so we can't return next hash.
	if len(hashes) < 1 {
		return tg.FileHash{}, false
	}

	// Sort hashes by offset.
	// Usually Telegram server returns sorted parts, but...
	// you never known what can they do.
	sort.SliceStable(hashes, func(i, j int) bool {
		return hashes[i].Offset < hashes[j].Offset
	})

	last := hashes[len(hashes)-1]
	// Check if we have reached the end.
	// If current state offset is equal the last offset + limit (right border)
	// then we got all hashes.
	if last.Offset == v.offset-last.Limit {
		return tg.FileHash{}, false
	}

	// Otherwise, we update current offset and add hashes to the end of queue.
	v.offset = last.Offset + last.Limit
	v.hashes = append(v.hashes, hashes...)
	return v.pop()
}

func (v *verifier) next(ctx context.Context) (tg.FileHash, bool, error) {
	v.mux.Lock()
	defer v.mux.Unlock()

	hash, ok := v.pop()
	if ok {
		return hash, ok, nil
	}

	for {
		hashes, err := v.client.Hashes(ctx, v.offset)
		if flood, err := tgerr.FloodWait(ctx, err); err != nil {
			if flood || tgerr.Is(err, tg.ErrTimeout) {
				continue
			}
			return tg.FileHash{}, false, xerrors.Errorf("get hashes: %w", err)
		}

		hash, ok = v.update(hashes...)
		return hash, ok, nil
	}
}

func (v *verifier) verify(hash tg.FileHash, data []byte) bool {
	return bytes.Equal(crypto.SHA256(data), hash.Hash)
}
