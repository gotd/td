package downloader

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"

	"github.com/gotd/td/tg"
)

// Builder is a download builder.
type Builder struct {
	downloader *Downloader

	schema schema
	hashes []tg.FileHash
	// verify controls legacy outer verifier (reader + verifier queue).
	// CDN redirect path has mandatory protocol-level verification in cdn schema,
	// independent from this flag.
	verify  bool
	threads int
	// scheduler overrides Downloader scheduler, if set.
	scheduler *Scheduler

	// offset is the offset to start download from (WithOffset).
	offset int64
	// size is the known file size, zero if unknown (WithSize).
	size int64
	// progress is called after each written part (WithProgress).
	progress ProgressFunc
}

// ProgressFunc is called after each written part with the offset of the end
// of downloaded data, counting from the start of the file (including the
// offset set by Builder.WithOffset), and the file size if known (see
// Builder.WithSize), zero otherwise.
//
// Calls are not concurrent.
type ProgressFunc func(done, total int64)

func newBuilder(downloader *Downloader, schema schema) *Builder {
	return &Builder{
		schema:     schema,
		threads:    1,
		downloader: downloader,
	}
}

// WithThreads sets downloading goroutines limit.
func (b *Builder) WithThreads(threads int) *Builder {
	if threads > 0 {
		b.threads = threads
	}
	return b
}

// WithRetryHandler sets callback for transient download errors that are retried
// internally.
//
// Handler can be called concurrently from download workers.
func (b *Builder) WithRetryHandler(handler RetryHandler) *Builder {
	switch s := b.schema.(type) {
	case master:
		s.retryHandler = handler
		b.schema = s
	case web:
		s.retryHandler = handler
		b.schema = s
	case *cdn:
		s.retryHandler = handler
	}

	return b
}

// WithVerify controls global hash verification behavior.
//
// `true` enables classic verifier reader (preloads hash queue and validates all
// chunks, both legacy and CDN).
// `false` disables classic verifier reader.
//
// If not called explicitly:
// - non-CDN path preserves old behavior (no upfront hash requests);
// - CDN path enables strict inline CDN verification after redirect.
//
// Use WithVerify(true) to force verifier-queue mode on all paths.
func (b *Builder) WithVerify(verify bool) *Builder {
	b.verify = verify
	return b
}

// WithOffset sets offset to start download from, e.g. to resume a partially
// downloaded file. Data before offset is not downloaded.
//
// Offset must be divisible by the part size (see Downloader.WithPartSize),
// unless verification is enabled via WithVerify(true), which only requires
// 4KB alignment: the hash window containing offset is downloaded whole to
// verify it, and data before offset is dropped.
//
// Stream writes data from offset to the start of io.Writer, Parallel writes
// it at the same offsets of io.WriterAt, and ToPath keeps existing file
// content before offset.
func (b *Builder) WithOffset(offset int64) *Builder {
	b.offset = offset
	return b
}

// WithSize sets known file size (e.g. Document.Size), so download stops
// requesting parts at the end of file instead of probing past it, and
// progress reports the total.
func (b *Builder) WithSize(size int64) *Builder {
	if size > 0 {
		b.size = size
	}
	return b
}

// WithProgress sets callback called after each written part.
func (b *Builder) WithProgress(progress ProgressFunc) *Builder {
	b.progress = progress
	return b
}

const offsetAlign = 4 * 1024

func (b *Builder) validate() error {
	if b.offset < 0 {
		return errors.Errorf("invalid offset %d", b.offset)
	}
	if b.offset%offsetAlign != 0 {
		return errors.Errorf("offset %d is not divisible by %d", b.offset, offsetAlign)
	}
	if !b.verify && b.downloader.partSize > 0 && b.offset%int64(b.downloader.partSize) != 0 {
		return errors.Errorf("offset %d is not divisible by part size %d", b.offset, b.downloader.partSize)
	}
	return nil
}

func (b *Builder) prepareMaster(m master, allowCDN bool) *Builder {
	clone := *b
	masterSchema := m
	// Keep explicit switch in schema to guarantee old request path when
	// CDN is disabled or unavailable.
	masterSchema.allowCDN = allowCDN
	clone.schema = masterSchema
	clone.hashes = nil
	return &clone
}

func (b *Builder) prepareCDNPath(m master, provider CDNProvider) *Builder {
	// Enable redirect errors on master schema (`upload.fileCdnRedirect`) while
	// still serving regular files from master when redirect is not required.
	m.allowCDN = true

	clone := *b
	// Avoid outer verifier on default path to keep non-redirect requests
	// equivalent to legacy master flow; CDN schema still verifies redirected
	// chunks inline according to Telegram CDN protocol.
	verifyCDNInline := !clone.verify
	clone.hashes = nil
	clone.schema = newCDNSchema(
		m,
		provider,
		b.downloader.pool,
		int64(b.threads),
		verifyCDNInline,
		m.retryHandler,
	)
	return &clone
}

func (b *Builder) shouldAllowCDN() bool {
	// CDN redirect flow is explicit-only to avoid hidden behavior changes and
	// preserve backwards compatibility for callers that never opted in.
	if b.downloader.allowCDN == nil {
		return false
	}
	return *b.downloader.allowCDN
}

func closeSchema(s schema) func() error {
	if closer, ok := s.(interface{ Close() error }); ok {
		return closer.Close
	}
	return nil
}

func (b *Builder) prepare() (_ *Builder, closeCDN func() error, err error) {
	if err := b.validate(); err != nil {
		return nil, nil, err
	}
	m, ok := b.schema.(master)
	if !ok {
		return b, closeSchema(b.schema), nil
	}

	// Fast path compatibility guarantee:
	// if CDN is not explicitly allowed we keep legacy master flow exactly as is,
	// without CDN pool creation and without extra hash requests.
	if !b.shouldAllowCDN() {
		prepared := b.prepareMaster(m, false)
		return prepared, closeSchema(prepared.schema), nil
	}

	// Even with AllowCDN=true, fallback to legacy master flow if client does not
	// provide CDN transport factory.
	provider, hasProvider := m.client.(CDNProvider)
	if !hasProvider {
		prepared := b.prepareMaster(m, false)
		return prepared, closeSchema(prepared.schema), nil
	}

	prepared := b.prepareCDNPath(m, provider)
	return prepared, closeSchema(prepared.schema), nil
}

func (b *Builder) reader() *reader {
	var r *reader
	if b.verify {
		v := newVerifier(b.schema, b.hashes...)
		v.start(b.offset)
		r = verifiedReader(b.schema, v)
		r.start = b.offset
	} else {
		r = plainReader(b.schema, b.downloader.partSize)
		r.offset = b.offset
	}
	r.size = b.size
	r.scheduler = b.downloader.scheduler
	if b.scheduler != nil {
		r.scheduler = b.scheduler
	}
	return r
}

func (b *Builder) sink() sink {
	return sink{
		done:     b.offset,
		total:    b.size,
		progress: b.progress,
	}
}

// Stream downloads file to given io.Writer.
// NB: in this mode download can't be parallel.
func (b *Builder) Stream(ctx context.Context, output io.Writer) (_ tg.StorageFileTypeClass, err error) {
	prepared, closeCDN, err := b.prepare()
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeCDN != nil {
			multierr.AppendInto(&err, closeCDN())
		}
	}()
	typ, runErr := prepared.downloader.stream(ctx, prepared.reader(), prepared.sink(), output)
	return typ, runErr
}

// Parallel downloads file to given io.WriterAt.
func (b *Builder) Parallel(ctx context.Context, output io.WriterAt) (_ tg.StorageFileTypeClass, err error) {
	prepared, closeCDN, err := b.prepare()
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeCDN != nil {
			multierr.AppendInto(&err, closeCDN())
		}
	}()

	typ, runErr := prepared.downloader.parallel(ctx, prepared.reader(), prepared.sink(), prepared.threads, output)
	return typ, runErr
}

// ToPath downloads file to given path.
//
// If offset is set (see WithOffset), existing file is not truncated, so data
// before offset is kept. If size is set (see WithSize), the file is truncated
// to it after a successful download.
func (b *Builder) ToPath(ctx context.Context, path string) (_ tg.StorageFileTypeClass, err error) {
	flag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	if b.offset > 0 {
		flag = os.O_RDWR | os.O_CREATE
	}
	f, err := os.OpenFile(filepath.Clean(path), flag, 0o666)
	if err != nil {
		return nil, errors.Wrap(err, "create output file")
	}
	defer func() {
		multierr.AppendInto(&err, f.Close())
	}()

	typ, runErr := b.Parallel(ctx, f)
	if runErr != nil {
		return typ, runErr
	}
	if b.size > 0 {
		// Drop stale data past the end of file, e.g. from a longer .part file.
		if err := f.Truncate(b.size); err != nil {
			return typ, errors.Wrap(err, "truncate output file")
		}
	}
	return typ, nil
}
