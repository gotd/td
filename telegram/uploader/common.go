package uploader

import (
	"context"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/mtproto"
)

func floodWait(ctx context.Context, err error) (bool, error) {
	var rpcErr *mtproto.Error
	if xerrors.As(err, &rpcErr) && rpcErr.Type == "FLOOD_WAIT" {
		select {
		case <-time.After(time.Duration(rpcErr.Argument) * time.Second):
			return true, err
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}

	return false, err
}
