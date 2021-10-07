package uploader

import (
	"context"
	"io"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
)

func (u *Uploader) smallLoop(ctx context.Context, h io.Writer, upload *Upload) error {
	buf := u.pool.GetSize(u.partSize)
	defer u.pool.Put(buf)

	last := false

	r := io.TeeReader(upload.from, h)
	for {
		n, err := io.ReadFull(r, buf.Buf)
		switch {
		case xerrors.Is(err, io.ErrUnexpectedEOF):
			last = true
		case xerrors.Is(err, io.EOF):
			return nil
		case err != nil:
			return xerrors.Errorf("read source: %w", err)
		}
		read := buf.Buf[:n]

		// Upload loop.
		for {
			r, err := u.rpc.UploadSaveFilePart(ctx, &tg.UploadSaveFilePartRequest{
				FileID:   upload.id,
				FilePart: int(upload.sentParts.Load()) % partsLimit,
				Bytes:    read,
			})

			if flood, err := tgerr.FloodWait(ctx, err); err != nil {
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
		if err := u.callback(ctx, upload.confirmSmall(n)); err != nil {
			return xerrors.Errorf("progress callback: %w", err)
		}

		if last {
			break
		}
	}

	return nil
}
