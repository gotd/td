package downloader

import (
	"context"
	"io"

	"github.com/gotd/td/tg"
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

// CDNProvider creates client bound to requested CDN DC.
// Returned closer is schema-scoped; for shared client-level pool adapters this
// can be a no-op closer.
type CDNProvider interface {
	CDN(ctx context.Context, dc int, max int64) (CDN, io.Closer, error)
}

type chunk struct {
	data []byte
	tag  tg.StorageFileTypeClass
}

// schema is simple interface for different download schemas.
type schema interface {
	Chunk(ctx context.Context, offset int64, limit int) (chunk, error)
	Hashes(ctx context.Context, offset int64) ([]tg.FileHash, error)
}
