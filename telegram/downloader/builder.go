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
}

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
	if b.verify {
		return verifiedReader(b.schema, newVerifier(b.schema, b.hashes...))
	}

	return plainReader(b.schema, b.downloader.partSize)
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
	typ, runErr := prepared.downloader.stream(ctx, prepared.reader(), output)
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

	typ, runErr := prepared.downloader.parallel(ctx, prepared.reader(), prepared.threads, output)
	return typ, runErr
}

// ToPath downloads file to given path.
func (b *Builder) ToPath(ctx context.Context, path string) (_ tg.StorageFileTypeClass, err error) {
	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return nil, errors.Wrap(err, "create output file")
	}
	defer func() {
		multierr.AppendInto(&err, f.Close())
	}()

	typ, runErr := b.Parallel(ctx, f)
	return typ, runErr
}
