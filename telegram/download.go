package telegram

import (
	"context"
	"io"
	"sync"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

// downloadClient is a downloader.Client that serves CDN redirects through
// CDN pools of the client.
type downloadClient struct {
	client *Client
	// raw is the invoker for master DC requests; nil means client.API().
	raw *tg.Client
}

func (d downloadClient) api() *tg.Client {
	if d.raw != nil {
		return d.raw
	}
	return d.client.API()
}

func (d downloadClient) UploadGetFile(
	ctx context.Context,
	request *tg.UploadGetFileRequest,
) (tg.UploadFileClass, error) {
	resp, err := d.api().UploadGetFile(ctx, request)

	return resp, err
}

func (d downloadClient) UploadGetFileHashes(
	ctx context.Context,
	request *tg.UploadGetFileHashesRequest,
) ([]tg.FileHash, error) {
	resp, err := d.api().UploadGetFileHashes(ctx, request)

	return resp, err
}

func (d downloadClient) UploadReuploadCDNFile(
	ctx context.Context,
	request *tg.UploadReuploadCDNFileRequest,
) ([]tg.FileHash, error) {
	resp, err := d.api().UploadReuploadCDNFile(ctx, request)

	return resp, err
}

func (d downloadClient) UploadGetCDNFileHashes(
	ctx context.Context,
	request *tg.UploadGetCDNFileHashesRequest,
) ([]tg.FileHash, error) {
	resp, err := d.api().UploadGetCDNFileHashes(ctx, request)

	return resp, err
}

func (d downloadClient) UploadGetWebFile(
	ctx context.Context,
	request *tg.UploadGetWebFileRequest,
) (*tg.UploadWebFile, error) {
	return d.api().UploadGetWebFile(ctx, request)
}

func (d downloadClient) CDN(ctx context.Context, dc int, max int64) (downloader.CDN, io.Closer, error) {
	return cdnProvider{client: d.client}.CDN(ctx, dc, max)
}

// cdnProvider implements downloader.CDNProvider using CDN pools of the client.
type cdnProvider struct {
	client *Client
}

func (p cdnProvider) CDN(ctx context.Context, dc int, max int64) (downloader.CDN, io.Closer, error) {
	invoker, err := p.client.CDN(ctx, dc, max)
	if err != nil {
		return nil, nil, err
	}
	if invoker == nil {
		return nil, nil, errors.New("telegram CDN pool returned nil invoker")
	}

	// CDN pools are cached on client level; downloader should not close them
	// by itself; lifecycle is controlled by caller via returned closer.
	cdnClient := tg.NewClient(invoker)

	return cdnClient, invoker, nil
}

// Downloader returns file downloader configured for current client: CDN
// policy from Options.AllowCDN.
//
// It has no Scheduler: use Downloader.WithScheduler with
// Client.DownloadScheduler of the DC the downloads go to.
func (c *Client) Downloader() *downloader.Downloader {
	return downloader.NewDownloader().WithAllowCDN(c.allowCDN)
}

// homeDownloader returns Downloader with Scheduler of the home DC, if
// Options.DownloadScheduler is set.
func (c *Client) homeDownloader() *downloader.Downloader {
	d := c.Downloader()
	if c.downloadSchedulers == nil || c.session == nil {
		return d
	}
	return d.WithScheduler(c.DownloadScheduler(c.session.Load().DC))
}

// DownloadScheduler returns the download Scheduler of given DC, created on
// first use with Options.DownloadScheduler, or nil if it is not set.
//
// Each DC has its own window: a flood wait of one DC doesn't shrink the
// window of another one.
func (c *Client) DownloadScheduler(dc int) *downloader.Scheduler {
	return c.downloadSchedulers.get(dc)
}

// downloadSchedulers lazily creates one download Scheduler per DC.
type downloadSchedulers struct {
	opts downloader.SchedulerOptions

	mux sync.Mutex
	m   map[int]*downloader.Scheduler
}

func newDownloadSchedulers(opts *downloader.SchedulerOptions) *downloadSchedulers {
	if opts == nil {
		return nil
	}
	return &downloadSchedulers{opts: *opts, m: map[int]*downloader.Scheduler{}}
}

func (d *downloadSchedulers) get(dc int) *downloader.Scheduler {
	if d == nil {
		return nil
	}
	d.mux.Lock()
	defer d.mux.Unlock()
	s, ok := d.m[dc]
	if !ok {
		s = downloader.NewScheduler(d.opts)
		d.m[dc] = s
	}
	return s
}

// CDNProvider returns downloader.CDNProvider that serves CDN redirects using
// cached CDN pools of the client.
//
// Wrap any downloader.Client with it via downloader.WithCDNProvider to enable
// CDN downloads through it.
func (c *Client) CDNProvider() downloader.CDNProvider {
	return cdnProvider{client: c}
}

// Download creates Builder for plain file downloads.
func (c *Client) Download(location tg.InputFileLocationClass) *downloader.Builder {
	return c.homeDownloader().Download(downloadClient{client: c}, location)
}

// DownloadVia creates Builder for plain file downloads, sending requests to
// master DC through given invoker, e.g. a pool from Client.Pool, Client.DC or
// Client.MediaOnly.
//
// Unlike passing tg.NewClient(invoker) to downloader directly, CDN redirects
// are served through CDN pools of the client (if CDN is allowed). No
// Scheduler is set, as the DC of invoker is unknown: use
// Builder.WithScheduler(c.DownloadScheduler(dc)).
func (c *Client) DownloadVia(invoker tg.Invoker, location tg.InputFileLocationClass) *downloader.Builder {
	return c.Downloader().Download(downloadClient{client: c, raw: tg.NewClient(invoker)}, location)
}

// DownloadWeb creates Builder for web file downloads.
func (c *Client) DownloadWeb(location tg.InputWebFileLocationClass) *downloader.Builder {
	return c.homeDownloader().Web(downloadClient{client: c}, location)
}

// DownloadWebVia creates Builder for web file downloads through given invoker.
func (c *Client) DownloadWebVia(invoker tg.Invoker, location tg.InputWebFileLocationClass) *downloader.Builder {
	return c.Downloader().Web(downloadClient{client: c, raw: tg.NewClient(invoker)}, location)
}
