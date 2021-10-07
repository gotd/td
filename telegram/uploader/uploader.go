package uploader

import (
	"context"
	"crypto/md5" // #nosec G501
	"encoding/hex"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/telegram/uploader/source"
	"github.com/nnqq/td/tg"
)

// Uploader is Telegram file uploader.
type Uploader struct {
	rpc      Client
	id       func() (int64, error)
	partSize int
	pool     *bin.Pool
	threads  int
	progress Progress
	src      source.Source
}

// NewUploader creates new Uploader.
func NewUploader(rpc Client) *Uploader {
	return (&Uploader{
		rpc: rpc,
		id: func() (int64, error) {
			return crypto.RandInt64(crypto.DefaultRand())
		},
		src:     source.NewHTTPSource(),
		threads: 1,
	}).WithPartSize(defaultPartSize)
}

// WithProgress sets progress callback.
func (u *Uploader) WithProgress(progress Progress) *Uploader {
	u.progress = progress
	return u
}

// WithSource sets URL resolver to use.
func (u *Uploader) WithSource(src source.Source) *Uploader {
	u.src = src
	return u
}

// WithThreads sets uploading goroutines limit per upload.
func (u *Uploader) WithThreads(threads int) *Uploader {
	if threads > 0 {
		u.threads = threads
	}
	return u
}

// WithIDGenerator sets id generator.
func (u *Uploader) WithIDGenerator(cb func() (int64, error)) *Uploader {
	u.id = cb
	return u
}

// WithPartSize sets part size.
// Should be divisible by 1024.
// 524288 should be divisible by partSize.
//
// See https://core.telegram.org/api/files#uploading-files.
func (u *Uploader) WithPartSize(partSize int) *Uploader {
	u.partSize = partSize
	u.pool = bin.NewPool(partSize)
	return u
}

// Upload uploads data from Upload object.
func (u *Uploader) Upload(ctx context.Context, upload *Upload) (tg.InputFileClass, error) {
	if err := checkPartSize(u.partSize); err != nil {
		return nil, xerrors.Errorf("invalid part size: %w", err)
	}

	if err := u.initUpload(upload); err != nil {
		return nil, err
	}

	if !upload.big {
		return u.uploadSmall(ctx, upload)
	}

	return u.uploadBig(ctx, upload)
}

func (u *Uploader) uploadSmall(ctx context.Context, upload *Upload) (tg.InputFileClass, error) {
	h := md5.New() // #nosec G401
	if err := u.smallLoop(ctx, h, upload); err != nil {
		return nil, err
	}

	return &tg.InputFile{
		ID:          upload.id,
		Parts:       int(upload.sentParts.Load()),
		Name:        upload.name,
		MD5Checksum: hex.EncodeToString(h.Sum(nil)),
	}, nil
}

func (u *Uploader) uploadBig(ctx context.Context, upload *Upload) (tg.InputFileClass, error) {
	if err := u.bigLoop(ctx, u.threads, upload); err != nil {
		return nil, err
	}

	return &tg.InputFileBig{
		ID:    upload.id,
		Parts: int(upload.sentParts.Load()),
		Name:  upload.name,
	}, nil
}

func (u *Uploader) callback(ctx context.Context, state ProgressState) error {
	if u.progress != nil {
		return u.progress.Chunk(ctx, state)
	}

	return nil
}
