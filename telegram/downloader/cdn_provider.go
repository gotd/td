package downloader

import (
	"context"
	"io"
)

// cdnClient is a Client extended with CDNProvider.
type cdnClient struct {
	Client
	provider CDNProvider
}

func (c cdnClient) CDN(ctx context.Context, dc int, max int64) (CDN, io.Closer, error) {
	return c.provider.CDN(ctx, dc, max)
}

// WithCDNProvider returns Client that serves CDN redirects through given
// provider, e.g. telegram.Client.CDNProvider.
//
// Without CDNProvider, downloads through plain clients (like
// tg.NewClient(pool)) silently fall back to master DC, even with
// Downloader.WithAllowCDN(true).
func WithCDNProvider(client Client, provider CDNProvider) Client {
	if provider == nil {
		return client
	}
	return cdnClient{Client: client, provider: provider}
}
