package telegram

import (
	"context"
	"io"
	"sync"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/gotd/log"

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
// policy from Options.AllowCDN, metrics from Options.MeterProvider and logs
// from Options.Logger (named "downloader").
//
// Each call returns a new copy, so its With* methods don't affect other
// downloads; instruments are shared and created once.
//
// It has no Scheduler: use Downloader.WithScheduler with
// Client.DownloadScheduler of the DC the downloads go to.
func (c *Client) Downloader() *downloader.Downloader {
	if c.downloader == nil {
		// Client is not created by NewClient.
		return downloader.NewDownloader().WithAllowCDN(c.allowCDN)
	}
	return c.downloader.Clone()
}

// newClientDownloader returns the instrumented base Downloader of a client.
func newClientDownloader(allowCDN bool, mp metric.MeterProvider, lg log.Logger) *downloader.Downloader {
	d := downloader.NewDownloader().WithAllowCDN(allowCDN)
	if mp != nil {
		d = d.WithMeterProvider(mp)
	}
	if lg != nil && lg != log.Nop {
		d = d.WithLogger(log.Named(lg, "downloader"))
	}
	return d
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

// downloaderMeterName is the meter of downloader metrics, the same as of
// downloader.Downloader.WithMeterProvider.
const downloaderMeterName = "github.com/gotd/td/telegram/downloader"

// downloadSchedulers lazily creates one download Scheduler per DC.
//
// With a meter, every Scheduler reports gotd.downloader.scheduler.* with
// attribute dc while the client runs.
type downloadSchedulers struct {
	opts  downloader.SchedulerOptions
	meter metric.Meter // nil if no metrics

	mux  sync.Mutex
	m    map[int]*downloader.Scheduler
	regs map[int]metric.Registration
}

func newDownloadSchedulers(opts *downloader.SchedulerOptions, mp metric.MeterProvider) *downloadSchedulers {
	if opts == nil {
		return nil
	}
	d := &downloadSchedulers{
		opts: *opts,
		m:    map[int]*downloader.Scheduler{},
		regs: map[int]metric.Registration{},
	}
	if mp != nil {
		d.meter = mp.Meter(downloaderMeterName)
	}
	return d
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
	if _, ok := d.regs[dc]; !ok && d.meter != nil {
		if reg, err := s.Instrument(d.meter, attribute.Int("dc", dc)); err == nil {
			d.regs[dc] = reg
		}
	}
	return s
}

// unregister stops reporting metrics of schedulers, e.g. when Run returns.
// Schedulers used again report them again.
func (d *downloadSchedulers) unregister() {
	if d == nil {
		return
	}
	d.mux.Lock()
	defer d.mux.Unlock()
	for dc, reg := range d.regs {
		_ = reg.Unregister()
		delete(d.regs, dc)
	}
}

// registered returns the number of schedulers reporting metrics.
func (d *downloadSchedulers) registered() int {
	if d == nil {
		return 0
	}
	d.mux.Lock()
	defer d.mux.Unlock()
	return len(d.regs)
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
