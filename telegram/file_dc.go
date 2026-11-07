package telegram

import (
	"context"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/attribute"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

const defaultFilePoolSize = 4

// FileDC returns invoker for file requests (e.g. upload.getFile) to given DC.
//
// Unlike Client.DC, the pool is created once per DC (concurrent callers share
// creation), cached and closed when Run returns, so callers can call FileDC
// for every request. Its size is Options.FilePoolSize.
//
// The invoker:
//   - is wrapped with Options.Middlewares (flood wait, rate limit, metrics);
//   - uses the primary session for the home DC, without authorization
//     transfer;
//   - follows FILE_MIGRATE_X errors to FileDC(X).
//
// Must be called while client is running, otherwise an error is returned.
//
// If a request fails as unauthorized, the cached pool of that DC is dropped
// and the request is retried once on a new pool with authorization transfer.
func (c *Client) FileDC(ctx context.Context, dc int) (tg.Invoker, error) {
	c.fileInvokersMux.Lock()
	inv, ok := c.fileInvokers[dc]
	c.fileInvokersMux.Unlock()
	if ok {
		return inv, nil
	}

	// Create pool eagerly to return creation errors.
	if _, err := c.cachedDC(ctx, c.filePools, dc); err != nil {
		return nil, errors.Wrapf(err, "create pool to DC %d", dc)
	}

	inv = chainMiddlewares(fileDCInvoker{client: c, dc: dc}, c.mw...)
	c.fileInvokersMux.Lock()
	if existing, ok := c.fileInvokers[dc]; ok {
		inv = existing
	} else {
		c.fileInvokers[dc] = inv
	}
	c.fileInvokersMux.Unlock()
	return inv, nil
}

// fileDCInvoker invokes requests on cached file pool of DC, following
// FILE_MIGRATE errors.
type fileDCInvoker struct {
	client *Client
	dc     int
}

// maxFileMigrations bounds FILE_MIGRATE hops of one request.
const maxFileMigrations = 2

func (f fileDCInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	dc := f.dc
	for hop := 0; ; hop++ {
		err := f.client.invokeCachedDC(ctx, f.client.filePools, dc, input, output)
		rpcErr, ok := tgerr.AsType(err, "FILE_MIGRATE")
		if !ok || hop >= maxFileMigrations || rpcErr.Argument == dc {
			return err
		}
		f.client.log.Debug(ctx, "Following FILE_MIGRATE",
			log.Int("from_dc", dc),
			log.Int("dc_id", rpcErr.Argument),
			log.Int("hop", hop+1),
		)
		f.client.meters().fileMigrations.Add(ctx, 1, dcAttrs(rpcErr.Argument,
			attribute.String("source", "file"),
		))
		dc = rpcErr.Argument
	}
}
