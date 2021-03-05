package downloader

import (
	"context"
	"io"
	"sync"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/tg"
)

func (d *Downloader) stream(ctx context.Context, r *reader, w io.Writer) (tg.StorageFileTypeClass, error) {
	var typ tg.StorageFileTypeClass
	typOnce := &sync.Once{}

	g := tdsync.NewCancellableGroup(ctx)
	toWrite := make(chan block, 1)
	// Download loop
	g.Go(func(ctx context.Context) error {
		for {
			b, err := r.Next(ctx)
			if err != nil {
				return xerrors.Errorf("get file: %w", err)
			}

			n := len(b.data)
			if n < 1 {
				typOnce.Do(func() {
					typ = b.tag
					close(toWrite)
				})
				return nil
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case toWrite <- b:
			}
		}
	})

	// Write loop
	g.Go(writeLoop(w, toWrite))

	return typ, g.Wait()
}
