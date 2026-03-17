package telegram

import (
	"context"
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

type downloadClient struct {
	client *Client
}

func (d downloadClient) api() *tg.Client {
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
	invoker, err := d.client.CDN(ctx, dc, max)
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

// Downloader returns file downloader configured for current client.
func (c *Client) Downloader() *downloader.Downloader {
	// Propagate explicit client-level CDN policy into downloader.
	return downloader.NewDownloader().WithAllowCDN(c.allowCDN)
}

// Download creates Builder for plain file downloads.
func (c *Client) Download(location tg.InputFileLocationClass) *downloader.Builder {
	return c.Downloader().Download(downloadClient{client: c}, location)
}

// DownloadWeb creates Builder for web file downloads.
func (c *Client) DownloadWeb(location tg.InputWebFileLocationClass) *downloader.Builder {
	return c.Downloader().Web(downloadClient{client: c}, location)
}
