package downloader

import (
	"context"
	"sync"
	"time"

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
	// scheduler limits in-flight requests, may be nil.
	scheduler *Scheduler // immutable

	// size is the known file size, zero if unknown.
	size int64
	// start is the offset to start download from (Builder.WithOffset).
	start int64

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
		observerFrom(ctx).hashMismatch(ctx, "verify", hash.Offset)
		return block{}, ErrHashMismatch
	}

	// The first hash window may start before the offset to start from (see
	// Builder.WithOffset): it is downloaded whole to verify it, but data
	// before the offset is dropped.
	if skip := r.start - b.offset; skip > 0 {
		if skip >= int64(len(b.data)) {
			b.data = nil
		} else {
			b.data = b.data[skip:]
		}
		b.offset = r.start
		if len(b.data) == 0 && !b.end {
			// Nothing left after the offset: the file ends before it.
			b.end = true
		}
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
		ch, err := r.chunk(ctx, offset, limit)

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

func (r *reader) chunk(ctx context.Context, offset int64, limit int) (chunk, error) {
	if r.scheduler == nil {
		return r.request(ctx, offset, limit)
	}
	epoch, err := r.scheduler.acquire(ctx)
	if err != nil {
		return chunk{}, err
	}
	rep := &backoffReporter{sch: r.scheduler, epoch: epoch}
	ch, err := r.request(withBackoffReporter(ctx, rep), offset, limit)
	if isSchedulerBackoff(ctx, err) {
		observerFrom(ctx).backoff(ctx, err)
	}
	rep.release(ctx, err)
	return ch, err
}

// request requests a part, reporting it to the observer of the download.
func (r *reader) request(ctx context.Context, offset int64, limit int) (chunk, error) {
	obs := observerFrom(ctx)
	if obs == nil {
		return r.sch.Chunk(ctx, offset, limit)
	}
	start := time.Now()
	ch, err := r.sch.Chunk(ctx, offset, limit)
	obs.request(ctx, requestKind(r.sch, ch, err), start, len(ch.data), err)
	return ch, err
}

// requestKind returns kind of the request of a part: master, cdn or web.
func requestKind(sch schema, ch chunk, err error) string {
	switch s := sch.(type) {
	case web:
		return kindWeb
	case *cdn:
		if ch.cdn {
			return kindCDN
		}
		if err != nil {
			if mode, _, _ := s.snapshot(); mode == modeCDN {
				return kindCDN
			}
		}
	}
	return kindMaster
}
