package message

import (
	"context"
	"io"
	"io/fs"

	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

type uploadBuilder struct {
	upload *uploader.Uploader
}

// UploadOption is a UploadBuilder creation option.
type UploadOption interface {
	apply(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error)
}

// uploadOptionFunc is a functional adapter for UploadOption.
type uploadOptionFunc func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error)

func (f uploadOptionFunc) apply(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error) {
	return f(ctx, b)
}

// FromFile uploads given File.
// NB: UploadFromFile does not close given file.
func FromFile(f uploader.File) UploadOption {
	return uploadOptionFunc(func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error) {
		return b.upload.FromFile(ctx, f)
	})
}

// FromPath uploads file from given path.
func FromPath(path string) UploadOption {
	return uploadOptionFunc(func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error) {
		return b.upload.FromPath(ctx, path)
	})
}

// FromFS uploads file from given path using given fs.FS.
func FromFS(filesystem fs.FS, path string) UploadOption {
	return uploadOptionFunc(func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error) {
		return b.upload.FromFS(ctx, filesystem, path)
	})
}

// FromReader uploads file from given io.Reader.
// NB: totally stream should not exceed the limit for
// small files (10 MB as docs says, may be a bit bigger).
func FromReader(name string, r io.Reader) UploadOption {
	return uploadOptionFunc(func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error) {
		return b.upload.FromReader(ctx, name, r)
	})
}

// FromBytes uploads file from given byte slice.
func FromBytes(name string, data []byte) UploadOption {
	return uploadOptionFunc(func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error) {
		return b.upload.FromBytes(ctx, name, data)
	})
}
