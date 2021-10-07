package uploader

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/uploader/source"
	"github.com/nnqq/td/tg"
)

// File is file abstraction.
type File interface {
	Stat() (os.FileInfo, error)
	io.Reader
}

// FromFile uploads given File.
// NB: FromFile does not close given file.
func (u *Uploader) FromFile(ctx context.Context, f File) (tg.InputFileClass, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, xerrors.Errorf("stat: %w", err)
	}

	return u.Upload(ctx, NewUpload(info.Name(), f, info.Size()))
}

// FromPath uploads file from given path.
func (u *Uploader) FromPath(ctx context.Context, path string) (tg.InputFileClass, error) {
	return u.FromFS(ctx, osFS{}, path)
}

type osFS struct{}

func (o osFS) Open(name string) (fs.File, error) {
	return os.Open(filepath.Clean(name))
}

// FromFS uploads file from fs using given path.
func (u *Uploader) FromFS(ctx context.Context, filesystem fs.FS, path string) (_ tg.InputFileClass, err error) {
	f, err := filesystem.Open(path)
	if err != nil {
		return nil, xerrors.Errorf("open: %w", err)
	}
	defer func() {
		multierr.AppendInto(&err, f.Close())
	}()

	return u.FromFile(ctx, f)
}

// FromReader uploads file from given io.Reader.
// NB: totally stream should not exceed the limit for
// small files (10 MB as docs says, may be a bit bigger).
func (u *Uploader) FromReader(ctx context.Context, name string, f io.Reader) (tg.InputFileClass, error) {
	return u.Upload(ctx, NewUpload(name, f, -1))
}

// FromBytes uploads file from given byte slice.
func (u *Uploader) FromBytes(ctx context.Context, name string, b []byte) (tg.InputFileClass, error) {
	return u.Upload(ctx, NewUpload(name, bytes.NewReader(b), int64(len(b))))
}

// FromURL uses given source to upload to Telegram.
func (u *Uploader) FromURL(ctx context.Context, rawURL string) (_ tg.InputFileClass, rerr error) {
	return u.FromSource(ctx, u.src, rawURL)
}

// FromSource uses given source and URL to fetch data and upload it to Telegram.
func (u *Uploader) FromSource(ctx context.Context, src source.Source, rawURL string) (_ tg.InputFileClass, rerr error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, xerrors.Errorf("parse url %q: %w", rawURL, err)
	}

	f, err := src.Open(ctx, parsed)
	if err != nil {
		return nil, xerrors.Errorf("open %q: %w", rawURL, err)
	}
	defer func() {
		multierr.AppendInto(&rerr, f.Close())
	}()

	name := f.Name()
	if name == "" {
		return nil, xerrors.Errorf("invalid name %q got from %q", name, rawURL)
	}

	size := f.Size()
	if size < 0 {
		size = -1
	}

	return u.Upload(ctx, NewUpload(f.Name(), f, size))
}
