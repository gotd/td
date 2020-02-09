package downloader

import (
	"context"

	"github.com/gotd/td/tg"
)

// Client represents Telegram RPC client.
type Client interface {
	UploadGetFile(ctx context.Context, request *tg.UploadGetFileRequest) (tg.UploadFileClass, error)
	UploadGetFileHashes(ctx context.Context, request *tg.UploadGetFileHashesRequest) ([]tg.FileHash, error)

	UploadReuploadCdnFile(ctx context.Context, request *tg.UploadReuploadCdnFileRequest) ([]tg.FileHash, error)

	UploadGetCdnFile(ctx context.Context, request *tg.UploadGetCdnFileRequest) (tg.UploadCdnFileClass, error)
	UploadGetCdnFileHashes(ctx context.Context, request *tg.UploadGetCdnFileHashesRequest) ([]tg.FileHash, error)

	UploadGetWebFile(ctx context.Context, request *tg.UploadGetWebFileRequest) (*tg.UploadWebFile, error)
}

type part struct {
	data []byte
	tag  tg.StorageFileTypeClass
}

// schema is simple interface for different download schemas.
type schema interface {
	Part(ctx context.Context, offset, limit int) (part, error)
	Hashes(ctx context.Context, offset int) ([]tg.FileHash, error)
}
