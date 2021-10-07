package downloader

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// Builder is a download builder.
type Builder struct {
	downloader *Downloader

	schema  schema
	hashes  []tg.FileHash
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

// WithVerify sets verify parameter.
// If verify is true, file hashes will be checked
// Verify is true by default for CDN downloads.
func (b *Builder) WithVerify(verify bool) *Builder {
	b.verify = verify
	return b
}

func (b *Builder) reader() *reader {
	if b.verify {
		return verifiedReader(b.schema, newVerifier(b.schema, b.hashes...))
	}

	return plainReader(b.schema, b.downloader.partSize)
}

// Stream downloads file to given io.Writer.
// NB: in this mode download can't be parallel.
func (b *Builder) Stream(ctx context.Context, output io.Writer) (tg.StorageFileTypeClass, error) {
	return b.downloader.stream(ctx, b.reader(), output)
}

// Parallel downloads file to given io.WriterAt.
func (b *Builder) Parallel(ctx context.Context, output io.WriterAt) (tg.StorageFileTypeClass, error) {
	return b.downloader.parallel(ctx, b.reader(), b.threads, output)
}

// ToPath downloads file to given path.
func (b *Builder) ToPath(ctx context.Context, path string) (_ tg.StorageFileTypeClass, err error) {
	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return nil, xerrors.Errorf("create output file: %w", err)
	}
	defer func() {
		multierr.AppendInto(&err, f.Close())
	}()

	return b.Parallel(ctx, f)
}
