// Package downloader contains downloading files helpers.
package downloader

import (
	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// Downloader is Telegram file downloader.
type Downloader struct {
	partSize int
	pool     *bin.Pool
	// allowCDN is tri-state:
	// nil  -> keep default downloader behavior (CDN disabled),
	// true -> allow redirect flow,
	// false-> force legacy master-only flow.
	allowCDN *bool
	// retryHandler observes transient downloader errors that are retried.
	retryHandler RetryHandler
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

// WithAllowCDN explicitly enables or disables CDN redirect flow.
//
// This flag is explicit: if it is not set, downloader keeps legacy
// master-DC-only behavior and does not attempt CDN redirect handling.
// Client integration (`telegram.Client.Downloader`) sets this option from
// `telegram.Options.AllowCDN`.
func (d *Downloader) WithAllowCDN(allow bool) *Downloader {
	d.allowCDN = &allow
	return d
}

// WithRetryHandler sets callback for transient download errors that are retried
// internally by downloader.
//
// Handler can be called concurrently from download workers.
func (d *Downloader) WithRetryHandler(handler RetryHandler) *Downloader {
	d.retryHandler = handler
	return d
}

// Download creates Builder for plain downloads.
func (d *Downloader) Download(rpc Client, location tg.InputFileLocationClass) *Builder {
	return newBuilder(d, master{
		client:       rpc,
		precise:      true,
		allowCDN:     false,
		retryHandler: d.retryHandler,
		location:     location,
	})
}

// Web creates Builder for web files downloads.
func (d *Downloader) Web(rpc Client, location tg.InputWebFileLocationClass) *Builder {
	return newBuilder(d, web{
		client:       rpc,
		retryHandler: d.retryHandler,
		location:     location,
	})
}
