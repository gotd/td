package downloader

import (
	"context"
	"io"
	"sync"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/tg"
)

func (d *Downloader) stream(ctx context.Context, rpc schema, w io.Writer) (tg.StorageFileTypeClass, error) {
	var typ tg.StorageFileTypeClass
	typOnce := &sync.Once{}

	grp := tdsync.NewCancellableGroup(ctx)
	toWrite := make(chan []byte, 1)
	// Download loop
	grp.Go(func(groupCtx context.Context) error {
		offset := 0

		for {
			p, err := rpc.Part(ctx, offset, d.partSize)
			if err != nil {
				return xerrors.Errorf("get file: %w", err)
			}

			n := len(p.data)
			if n < 1 {
				typOnce.Do(func() {
					typ = p.tag
					close(toWrite)
				})
				return nil
			}

			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
			case toWrite <- p.data:
			}

			if n < d.partSize {
				typOnce.Do(func() {
					typ = p.tag
					close(toWrite)
				})
				return nil
			}

			offset += n
		}
	})

	// Write loop
	grp.Go(func(groupCtx context.Context) error {
		for {
			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
			case part, ok := <-toWrite:
				if !ok {
					return nil
				}

				_, err := w.Write(part)
				if err != nil {
					return xerrors.Errorf("write output: %w", err)
				}
			}
		}
	})

	return typ, grp.Wait()
}
