package message

import (
	"context"
	"io"
	"io/fs"

	"go.uber.org/atomic"

	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

// Uploader is a abstraction for Telegram file uploader.
type Uploader interface {
	FromFile(ctx context.Context, f uploader.File) (tg.InputFileClass, error)
	FromPath(ctx context.Context, path string) (tg.InputFileClass, error)
	FromFS(ctx context.Context, filesystem fs.FS, path string) (tg.InputFileClass, error)
	FromReader(ctx context.Context, name string, f io.Reader) (tg.InputFileClass, error)
	FromBytes(ctx context.Context, name string, b []byte) (tg.InputFileClass, error)
}

type uploadBuilder struct {
	upload Uploader
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

type fileCache atomic.Value

func (r *fileCache) Store(result tg.InputFileClass) {
	r.Value.Store(result)
}

func (r *fileCache) Load() (result tg.InputFileClass, ok bool) {
	result, ok = r.Value.Load().(tg.InputFileClass)
	return
}

func uploadOption(promise func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error)) UploadOption {
	once := &fileCache{}
	return uploadOptionFunc(func(ctx context.Context, b uploadBuilder) (r tg.InputFileClass, err error) {
		if v, ok := once.Load(); ok {
			return v, nil
		}
		defer func() {
			if err == nil && r != nil {
				once.Store(r)
			}
		}()

		return promise(ctx, b)
	})
}

// FromFile uploads given File.
// NB: FromFile does not close given file.
func FromFile(f uploader.File) UploadOption {
	return uploadOption(func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error) {
		return b.upload.FromFile(ctx, f)
	})
}

// FromPath uploads file from given path.
func FromPath(path string) UploadOption {
	return uploadOption(func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error) {
		return b.upload.FromPath(ctx, path)
	})
}

// FromFS uploads file from given path using given fs.FS.
func FromFS(filesystem fs.FS, path string) UploadOption {
	return uploadOption(func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error) {
		return b.upload.FromFS(ctx, filesystem, path)
	})
}

// FromReader uploads file from given io.Reader.
// NB: totally stream should not exceed the limit for
// small files (10 MB as docs says, may be a bit bigger).
func FromReader(name string, r io.Reader) UploadOption {
	return uploadOption(func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error) {
		return b.upload.FromReader(ctx, name, r)
	})
}

// FromBytes uploads file from given byte slice.
func FromBytes(name string, data []byte) UploadOption {
	return uploadOption(func(ctx context.Context, b uploadBuilder) (tg.InputFileClass, error) {
		return b.upload.FromBytes(ctx, name, data)
	})
}
