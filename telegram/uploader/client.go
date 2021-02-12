package uploader

import (
	"context"

	"github.com/gotd/td/tg"
)

// Client represents Telegram RPC client.
type Client interface {
	UploadSaveFilePart(ctx context.Context, request *tg.UploadSaveFilePartRequest) (bool, error)
	UploadSaveBigFilePart(ctx context.Context, request *tg.UploadSaveBigFilePartRequest) (bool, error)
}
