package uploader

import (
	"context"
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

func (u *Uploader) smallLoop(ctx context.Context, h io.Writer, upload *Upload) error {
	buf := u.pool.GetSize(u.partSize)
	defer u.pool.Put(buf)

	last := false

	r := io.LimitReader(upload.from, bigFileLimit)
	r = io.TeeReader(r, h)
	for {
		n, err := io.ReadFull(r, buf)
		switch {
		case xerrors.Is(err, io.ErrUnexpectedEOF):
			last = true
		case xerrors.Is(err, io.EOF):
			return nil
		case err != nil:
			return xerrors.Errorf("read source: %w", err)
		}
		read := buf[:n]

		// Upload loop.
		for {
			r, err := u.rpc.UploadSaveFilePart(ctx, &tg.UploadSaveFilePartRequest{
				FileID:   upload.id,
				FilePart: int(upload.sentParts.Load()) % partsLimit,
				Bytes:    read,
			})

			if flood, err := floodWait(ctx, err); err != nil {
				if flood {
					continue
				}
				return xerrors.Errorf("send upload RPC: %w", err)
			}

			// If Telegram returned false, it seems save is not successful, so we retry to send.
			if !r {
				continue
			}

			break
		}

		upload.sentParts.Inc()
		uploaded, parts := upload.confirm(n)
		if err := u.callback(ctx, ProgressState{
			Part:     parts,
			PartSize: u.partSize,
			Uploaded: uploaded,
			Total:    int(upload.totalBytes),
		}); err != nil {
			return xerrors.Errorf("progress callback: %w", err)
		}

		if last {
			break
		}
	}

	return nil
}
