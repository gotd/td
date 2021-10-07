// Package downloader contains downloading files helpers.
package downloader

import (
	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
)

// Downloader is Telegram file downloader.
type Downloader struct {
	partSize int
	pool     *bin.Pool
}

const defaultPartSize = 512 * 1024 // 512 kb

// NewDownloader creates new Downloader.
func NewDownloader() *Downloader {
	return new(Downloader).WithPartSize(defaultPartSize)
}

// WithPartSize sets chunk size.
// Must be divisible by 4KB.
//
// See https://core.telegram.org/api/files#downloading-files.
func (d *Downloader) WithPartSize(partSize int) *Downloader {
	d.partSize = partSize
	d.pool = bin.NewPool(partSize)
	return d
}

// Download creates Builder for plain downloads.
//
// Method sets CDNSupported field for upload.getFile. Use DownloadDirect for
// direct downloads without CDN support.
//
// See https://core.telegram.org/cdn.
func (d *Downloader) Download(rpc Client, location tg.InputFileLocationClass) *Builder {
	return newBuilder(d, master{
		client:   rpc,
		precise:  true,
		allowCDN: true,
		location: location,
	})
}

// DownloadDirect creates Builder for plain downloads with disabled CDN redirect
// handling.
func (d *Downloader) DownloadDirect(rpc Client, location tg.InputFileLocationClass) *Builder {
	return newBuilder(d, master{
		client:   rpc,
		precise:  true,
		allowCDN: false,
		location: location,
	})
}

// CDN creates Builder for CDN downloads.
func (d *Downloader) CDN(rpc Client, cdnRPC CDN, redirect *tg.UploadFileCDNRedirect) *Builder {
	b := newBuilder(d, cdn{
		cdn:      cdnRPC,
		client:   rpc,
		pool:     d.pool,
		redirect: redirect,
	})
	b.hashes = append(b.hashes, redirect.FileHashes...)
	return b
}

// Web creates Builder for web files downloads.
func (d *Downloader) Web(rpc Client, location tg.InputWebFileLocationClass) *Builder {
	return newBuilder(d, web{
		client:   rpc,
		location: location,
	})
}
