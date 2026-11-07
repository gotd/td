package downloader

import (
	"context"
	"io"
	"sync"

	"github.com/go-faster/errors"

	"github.com/gotd/td/syncio"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tg"
)

// nolint:gocognit
func (d *Downloader) parallel(
	ctx context.Context, r *reader, s sink,
	threads int, w io.WriterAt,
) (tg.StorageFileTypeClass, error) {
	var (
		typ    tg.StorageFileTypeClass
		typMux sync.Mutex
	)

	ready := tdsync.NewReady()
	g := tdsync.NewCancellableGroup(ctx)
	toWrite := make(chan block, threads)

	stop := func(t tg.StorageFileTypeClass) {
		// Blocks past known end of file (see Builder.WithSize) carry no type:
		// keep the first type reported by a server response.
		if t != nil {
			typMux.Lock()
			if typ == nil {
				typ = t
			}
			typMux.Unlock()
		}
		ready.Signal()
	}

	// Download loop
	g.Go(func(ctx context.Context) error {
		downloads := tdsync.NewCancellableGroup(ctx)
		defer close(toWrite)

		for i := 0; i < threads; i++ {
			downloads.Go(func(ctx context.Context) error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-ready.Ready():
						return nil
					default:
					}

					b, err := r.Next(ctx)
					if err != nil {
						return errors.Wrap(err, "get file")
					}

					// If returned chunk is zero, that means we read all file.
					n := len(b.data)
					if n < 1 {
						stop(b.tag)
						return nil
					}

					select {
					case <-ctx.Done():
						return ctx.Err()
					case toWrite <- b:
					}

					if b.last() {
						stop(b.tag)
						return nil
					}
				}
			})
		}

		return downloads.Wait()
	})

	// Write loop
	g.Go(writeAtLoop(syncio.NewWriterAt(w), s, toWrite))

	err := g.Wait()
	typMux.Lock()
	defer typMux.Unlock()
	return typ, err
}
