package downloader

import (
	"context"
	"io"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/tdsync"
	"github.com/nnqq/td/tg"
)

func (d *Downloader) stream(ctx context.Context, r *reader, w io.Writer) (tg.StorageFileTypeClass, error) {
	var typ tg.StorageFileTypeClass

	g := tdsync.NewCancellableGroup(ctx)
	toWrite := make(chan block, 1)

	stop := func(t tg.StorageFileTypeClass) {
		typ = t
		close(toWrite)
	}
	// Download loop
	g.Go(func(ctx context.Context) error {
		for {
			b, err := r.Next(ctx)
			if err != nil {
				return xerrors.Errorf("get file: %w", err)
			}

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

	// Write loop
	g.Go(writeLoop(w, toWrite))

	return typ, g.Wait()
}
