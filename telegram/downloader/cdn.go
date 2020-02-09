package downloader

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// ExpiredTokenError error is returned when Downloader get expired file token for CDN.
// See https://core.telegram.org/constructor/upload.fileCdnRedirect.
type ExpiredTokenError struct {
	*tg.UploadCdnFileReuploadNeeded
}

// Error implements error interface.
func (r *ExpiredTokenError) Error() string {
	return "redirect to master DC for requesting new file token"
}

// cdn is a CDN DC download schema.
// See https://core.telegram.org/cdn#getting-files-from-a-cdn.
type cdn struct {
	client   Client
	pool     *bin.Pool
	redirect *tg.UploadFileCdnRedirect
}

var _ schema = cdn{}

// decrypt decrypts file part from Telegram CDN.
// See https://core.telegram.org/cdn#decrypting-files.
func (c cdn) decrypt(src []byte, offset int) ([]byte, error) {
	block, err := aes.NewCipher(c.redirect.EncryptionKey)
	if err != nil {
		return nil, xerrors.Errorf("create cipher: %w", err)
	}

	// Copy IV to buffer from Pool.
	iv := c.pool.GetSize(len(c.redirect.EncryptionIv))
	defer c.pool.Put(iv)
	copy(iv.Buf, c.redirect.EncryptionIv)

	// For IV, it should use the value of encryption_iv, modified in the following manner:
	// for each offset replace the last 4 bytes of the encryption_iv with offset / 16 in big-endian.
	binary.BigEndian.PutUint32(iv.Buf[iv.Len()-4:], uint32(offset/16))

	dst := make([]byte, len(src))
	cipher.NewCTR(block, iv.Buf).XORKeyStream(dst, src)
	return dst, nil
}

func (c cdn) Part(ctx context.Context, offset, limit int) (part, error) {
	r, err := c.client.UploadGetCdnFile(ctx, &tg.UploadGetCdnFileRequest{
		Offset:    offset,
		Limit:     limit,
		FileToken: c.redirect.FileToken,
	})
	if err != nil {
		return part{}, err
	}

	switch result := r.(type) {
	case *tg.UploadCdnFile:
		data, err := c.decrypt(result.Bytes, offset)
		if err != nil {
			return part{}, err
		}

		return part{
			data: data,
		}, nil
	case *tg.UploadCdnFileReuploadNeeded:
		return part{}, &ExpiredTokenError{UploadCdnFileReuploadNeeded: result}
	default:
		return part{}, xerrors.Errorf("unexpected type %T", r)
	}
}

func (c cdn) Hashes(ctx context.Context, offset int) ([]tg.FileHash, error) {
	return c.client.UploadGetCdnFileHashes(ctx, &tg.UploadGetCdnFileHashesRequest{
		FileToken: c.redirect.FileToken,
		Offset:    offset,
	})
}
