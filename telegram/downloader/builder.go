package downloader

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// Builder is a download builder.
type Builder struct {
	downloader *Downloader

	schema  schema
	threads int
}

func newBuilder(downloader *Downloader, schema schema) *Builder {
	return &Builder{
		schema:     schema,
		threads:    runtime.GOMAXPROCS(0) * 2,
		downloader: downloader,
	}
}

// WithThreads sets downloading goroutines limit.
func (b *Builder) WithThreads(threads int) *Builder {
	b.threads = threads
	return b
}

// Stream downloads file to given io.Writer.
// NB: in this mode download can't be parallel.
func (b *Builder) Stream(ctx context.Context, output io.Writer) (tg.StorageFileTypeClass, error) {
	return b.downloader.stream(ctx, b.schema, output)
}

// Parallel downloads file to given io.WriterAt.
func (b *Builder) Parallel(ctx context.Context, output io.WriterAt) (tg.StorageFileTypeClass, error) {
	return b.downloader.parallel(ctx, b.schema, b.threads, output)
}

// ToPath downloads file to given path.
func (b *Builder) ToPath(ctx context.Context, path string) (tg.StorageFileTypeClass, error) {
	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return nil, xerrors.Errorf("create output file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	return b.Parallel(ctx, f)
}
