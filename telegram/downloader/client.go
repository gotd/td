package downloader

import (
	"context"

	"github.com/nnqq/td/tg"
)

// CDN represents Telegram RPC client to CDN server.
type CDN interface {
	UploadGetCDNFile(ctx context.Context, request *tg.UploadGetCDNFileRequest) (tg.UploadCDNFileClass, error)
}

// Client represents Telegram RPC client.
type Client interface {
	UploadGetFile(ctx context.Context, request *tg.UploadGetFileRequest) (tg.UploadFileClass, error)
	UploadGetFileHashes(ctx context.Context, request *tg.UploadGetFileHashesRequest) ([]tg.FileHash, error)

	UploadReuploadCDNFile(ctx context.Context, request *tg.UploadReuploadCDNFileRequest) ([]tg.FileHash, error)
	UploadGetCDNFileHashes(ctx context.Context, request *tg.UploadGetCDNFileHashesRequest) ([]tg.FileHash, error)

	UploadGetWebFile(ctx context.Context, request *tg.UploadGetWebFileRequest) (*tg.UploadWebFile, error)
}

type chunk struct {
	data []byte
	tag  tg.StorageFileTypeClass
}

// schema is simple interface for different download schemas.
type schema interface {
	Chunk(ctx context.Context, offset, limit int) (chunk, error)
	Hashes(ctx context.Context, offset int) ([]tg.FileHash, error)
}
