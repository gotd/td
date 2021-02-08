package downloader

import (
	"context"
	"strconv"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// Downloader is Telegram file downloader.
type Downloader struct {
	partSize int
	pool     *bin.Pool

	allowCDN bool
}

const (
	defaultPartSize = 1024 * 1024
)

// NewDownloader creates new Downloader.
func NewDownloader() *Downloader {
	return (&Downloader{
		allowCDN: false,
	}).WithPartSize(defaultPartSize)
}

// WithPartSize sets part size.
// Must be divisible by 4KB.
//
// See https://core.telegram.org/api/files#downloading-files.
func (d *Downloader) WithPartSize(partSize int) *Downloader {
	d.partSize = partSize
	d.pool = bin.NewPool(partSize)
	return d
}

// RedirectError error is returned when Downloader get CDN redirect.
// See https://core.telegram.org/constructor/upload.fileCdnRedirect.
type RedirectError struct {
	*tg.UploadFileCdnRedirect
}

// Error implements error interface.
func (r *RedirectError) Error() string {
	return "redirect to CDN DC " + strconv.Itoa(r.DCID)
}

// Download download data from Telegram server to given output.
func (d *Downloader) Download(ctx context.Context, rpc *tg.Client, download Download) (tg.StorageFileTypeClass, error) {
	if !download.cdn {
		return d.download(ctx, rpc, download)
	}

	return nil, xerrors.New("CDN downloads not implemented yet")
}

func (d *Downloader) download(ctx context.Context, rpc *tg.Client, download Download) (tg.StorageFileTypeClass, error) {
	offset := 0

	for {
		req := &tg.UploadGetFileRequest{
			Offset:   offset,
			Limit:    d.partSize,
			Location: download.file,
		}
		req.SetCDNSupported(d.allowCDN)

		f, err := rpc.UploadGetFile(ctx, req)
		if err != nil {
			return nil, xerrors.Errorf("get file: %w", err)
		}

		switch file := f.(type) {
		case *tg.UploadFile:
			if len(file.Bytes) < 1 {
				return file.Type, nil
			}

			n, err := download.output.Write(file.Bytes)
			if err != nil {
				return nil, xerrors.Errorf("write output: %w", err)
			}

			if n < d.partSize {
				return file.Type, nil
			}

			offset += n
		case *tg.UploadFileCdnRedirect:
			return nil, &RedirectError{
				UploadFileCdnRedirect: file,
			}
		}
	}
}
