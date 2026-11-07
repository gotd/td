package downloader

import (
	"context"
	"sync"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tgerr"
)

type block struct {
	chunk
	offset   int64
	partSize int
	// end is true if block reaches known end of file.
	end bool
}

// last compares partSize and chunk length to determine last part.
func (b block) last() bool {
	if b.end {
		return true
	}
	// If returned chunk is smaller than requested part, it seems
	// it is last part.
	return len(b.data) < b.partSize
}

type reader struct {
	sch      schema    // immutable
	verifier *verifier // immutable
	partSize int       // immutable

	// size is the known file size, zero if unknown.
	size int64

	offset    int64
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
	r.offset += int64(r.partSize)
	r.offsetMux.Unlock()

	if r.size > 0 && offset >= r.size {
		// Known end of file: do not request past it.
		return block{offset: offset, partSize: r.partSize}, nil
	}

	return r.next(ctx, offset, r.partSize)
}

func (r *reader) next(ctx context.Context, offset int64, limit int) (block, error) {
	retryAttempt := 0
	for {
		ch, err := r.sch.Chunk(ctx, offset, limit)

		if flood, err := tgerr.FloodWait(ctx, err); err != nil {
			if flood || isRetryableTimeout(ctx, err) {
				retryAttempt++
				reportSchemaRetry(r.sch, RetryOperationReaderChunk, retryAttempt, err)
				continue
			}
			return block{}, errors.Wrap(err, "get next chunk")
		}

		b := block{
			chunk:    ch,
			offset:   offset,
			partSize: r.partSize,
		}
		if r.size > 0 && offset+int64(len(ch.data)) >= r.size {
			b.end = true
		}
		return b, nil
	}
}
