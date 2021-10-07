package downloader

import (
	"context"
	"sync"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
)

type block struct {
	chunk
	offset   int64
	partSize int
}

// last compares partSize and chunk length to determine last part.
func (b block) last() bool {
	// If returned chunk is smaller than requested part, it seems
	// it is last part.
	return len(b.data) < b.partSize
}

type reader struct {
	sch      schema    // immutable
	verifier *verifier // immutable
	partSize int       // immutable

	offset    int
	offsetMux sync.Mutex
}

func verifiedReader(sch schema, verifier *verifier) *reader {
	return &reader{
		sch:      sch,
		verifier: verifier,
	}
}

func plainReader(sch schema, partSize int) *reader {
	return &reader{
		sch:      sch,
		partSize: partSize,
	}
}

func (r *reader) Next(ctx context.Context) (block, error) {
	if r.verifier != nil {
		return r.nextHashed(ctx)
	}

	return r.nextPlain(ctx)
}

func (r *reader) nextHashed(ctx context.Context) (block, error) {
	// Fetch next hashes.
	hash, ok, err := r.verifier.next(ctx)
	if err != nil {
		return block{}, err
	}
	if !ok {
		return block{}, nil
	}

	// Get next chunk.
	b, err := r.next(ctx, hash.Offset, hash.Limit)
	if err != nil {
		return block{}, err
	}

	// Verify chunk.
	if !r.verifier.verify(hash, b.data) {
		return block{}, ErrHashMismatch
	}

	return b, nil
}

func (r *reader) nextPlain(ctx context.Context) (block, error) {
	r.offsetMux.Lock()
	offset := r.offset
	r.offset += r.partSize
	r.offsetMux.Unlock()

	return r.next(ctx, offset, r.partSize)
}

func (r *reader) next(ctx context.Context, offset, limit int) (block, error) {
	for {
		ch, err := r.sch.Chunk(ctx, offset, limit)

		if flood, err := tgerr.FloodWait(ctx, err); err != nil {
			if flood || tgerr.Is(err, tg.ErrTimeout) {
				continue
			}
			return block{}, xerrors.Errorf("get next chunk: %w", err)
		}

		return block{
			chunk:    ch,
			offset:   int64(offset),
			partSize: r.partSize,
		}, nil
	}
}
