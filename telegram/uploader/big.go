package uploader

import (
	"context"
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/tg"
)

type part struct {
	id     int
	buf    *bin.Buffer
	upload *Upload
}

func (u *Uploader) uploadBigFilePart(ctx context.Context, p part) (int, error) {
	defer u.pool.Put(p.buf)

	// Upload loop.
	for {
		r, err := u.rpc.UploadSaveBigFilePart(ctx, &tg.UploadSaveBigFilePartRequest{
			FileID:         p.upload.id,
			FilePart:       p.id,
			FileTotalParts: p.upload.totalParts,
			Bytes:          p.buf.Buf,
		})

		if flood, err := floodWait(ctx, err); err != nil {
			if flood {
				continue
			}
			return 0, xerrors.Errorf("send upload part %d RPC: %w", p.id, err)
		}

		// If Telegram returned false, it seems save is not successful, so we retry to send.
		if r {
			return p.buf.Len(), nil
		}
	}
}

func (u *Uploader) bigLoop(ctx context.Context, threads int, upload *Upload) error { // nolint:gocognit
	grp := tdsync.NewCancellableGroup(ctx)
	toSend := make(chan part, threads)

	// Run read loop
	r := &syncReader{r: upload.from}
	grp.Go(func(groupCtx context.Context) error {
		last := false

		for {
			buf := u.pool.GetSize(u.partSize)

			n, err := io.ReadFull(r, buf.Buf)
			switch {
			case xerrors.Is(err, io.ErrUnexpectedEOF):
				last = true
			case xerrors.Is(err, io.EOF):
				u.pool.Put(buf)

				close(toSend)
				return nil
			case err != nil:
				u.pool.Put(buf)

				return xerrors.Errorf("read source: %w", err)
			}

			buf.Buf = buf.Buf[:n]
			nextPart := part{
				id:     int(upload.sentParts.Load()) % partsLimit,
				buf:    buf,
				upload: upload,
			}
			select {
			case toSend <- nextPart:
				upload.sentParts.Inc()
				if last {
					close(toSend)
					return nil
				}
			case <-groupCtx.Done():
				u.pool.Put(buf)

				return groupCtx.Err()
			}
		}
	})

	for i := 0; i < threads; i++ {
		grp.Go(func(groupCtx context.Context) error {
			for {
				select {
				case <-groupCtx.Done():
					return groupCtx.Err()
				case part, ok := <-toSend:
					if !ok {
						return nil
					}

					n, err := u.uploadBigFilePart(groupCtx, part)
					if err != nil {
						return xerrors.Errorf("upload part: %w", err)
					}

					uploaded, _ := upload.confirm(n)
					if err := u.callback(ctx, ProgressState{
						Part:     part.id,
						PartSize: u.partSize,
						Uploaded: uploaded,
						Total:    int(upload.totalBytes),
					}); err != nil {
						return xerrors.Errorf("progress callback: %w", err)
					}
				}
			}
		})
	}

	return grp.Wait()
}
